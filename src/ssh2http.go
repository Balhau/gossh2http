package main

import (
  "fmt"
  "strings"
  "encoding/base64"
  "github.com/fatih/color"
	"github.com/urfave/cli"
	"net"
	"os"
	//"io"
  "os/signal"
	"syscall"
	//"runtime"
)

func chanFromConn(conn net.Conn) chan []byte {
    c := make(chan []byte)
    go func() {
        b := make([]byte, 1024)
        for {
            n, err := conn.Read(b)
            if n > 0 {
                res := make([]byte, n)
                // Copy the buffer so it doesn't get changed while read by the recipient.
                copy(res, b[:n])
                c <- res
            }
            if err != nil {
                c <- nil
                break
            }
        }
    }()
    return c
}

func Pipe(conn1 net.Conn, conn2 net.Conn) {
    chan1 := chanFromConn(conn1)
    chan2 := chanFromConn(conn2)

    for {
        select {
        case b1 := <-chan1:
            if b1 == nil {
                return
            } else {
                conn2.Write(b1)
            }
        case b2 := <-chan2:
            if b2 == nil {
                return
            } else {
                conn1.Write(b2)
            }
        }
    }
}

func envelopeSSLServerHandshake(data []byte) string{
  base64sslHandShake := base64.StdEncoding.EncodeToString(data)
  httpEnvelope := `
      HTTP /1.1 200

      <body>{body}</body>
  `

  envelope := strings.Replace(httpEnvelope,"{body}",base64sslHandShake ,2)
  return envelope
}

func handleSshClientConnection(remoteAddress string,client net.Conn){

  bufIn := make([]byte, 1024)

  sshServer, err := net.Dial("tcp", remoteAddress)

  if err != nil {
    fmt.Println("Error connecting:", err.Error())
  }

  _,err = sshServer.Read(bufIn)

  if err != nil {
    fmt.Println("Error reading:", err.Error())
  }

  envelope := envelopeSSLServerHandshake(bufIn)

  client.Write([]byte(envelope))

  if err != nil {
    fmt.Println("Error reading:", err.Error())
  }

  client.Write([]byte(envelope))

  Pipe(sshServer,client)



  // Close the connection when you're done with it.

}

func ctrlc() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		color.Set(color.FgGreen)
		fmt.Println("\nExecution stopped by", sig)
		color.Unset()
		os.Exit(0)
	}()
}

func serve(remoteAddress string,localPort string){

  ln, _ := net.Listen("tcp","localhost:"+localPort)

  defer ln.Close()

  fmt.Println("Listening on :" + localPort)

  for{
    conn, _ := ln.Accept()
    go handleSshClientConnection(remoteAddress,conn)
  }
}


func main(){
  app := cli.NewApp()
	app.Name = "ssh2http"
	app.Version = "1.0.0"
	app.Usage = "Ssh to http packet wrapping"
	app.UsageText = "ssh2http --from <local_ssh2http_ip>:<port --to <remote_ssh2http_tunnel>:<port>"
	app.Copyright = "MIT License"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Balhau",
			Email: "balhau@balhau.net",
		},
	}

app.Flags = []cli.Flag{
  cli.BoolFlag{
			Name:  "serve, s",
			Usage: "list local addresses",
		},
  }

  color.Set(color.FgGreen)
  fmt.Println("This is a green message")
  color.Unset()
  app.Action = func(c *cli.Context) error {
    if c.Bool("serve"){

    }else{
      serve("localhost:10100","10000")
    }
    return nil
  }

  app.Run(os.Args)

}
