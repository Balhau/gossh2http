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

func extractBase64Payload(httpString string) string {
  strs0 := strings.Split(httpString,"<body>")
  strs1 := strings.Split(strs0[1],"</body>")
  return strs1[0]
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

func handleSshHandshakeServer(remoteAddress string,client net.Conn) (sshServer net.Conn){
  bufIn := make([]byte, 1024)

  sshServer, err := net.Dial("tcp", remoteAddress)

  if err != nil {
    fmt.Println("Error connecting:", err.Error())
  }

  fmt.Println("Reading payload")

  _,err = client.Read(bufIn)

  fmt.Println("Payload readed")

  if err != nil {
    fmt.Println("Error reading:", err.Error())
  }

  fmt.Println("Input String: ",string(bufIn))

  sshB64Payload := extractBase64Payload(string(bufIn))

  fmt.Println("sshb64Payload: ",sshB64Payload)

  payload, err := base64.StdEncoding.DecodeString(sshB64Payload)

  strPayload := strings.Trim(string(payload),"")
  fmt.Println("payload: ",strPayload)


  if err != nil {
    fmt.Println("Error reading:", err.Error())
  }

  sshServer.Write([]byte(strPayload))

  return sshServer
}

func handleSshHandshakeClient(remoteAddress string,client net.Conn) (sshServer net.Conn){
  bufIn := make([]byte, 1024)

  sshServer, err := net.Dial("tcp", remoteAddress)

  if err != nil {
    fmt.Println("Error connecting:", err.Error())
  }

  _,err = client.Read(bufIn)

  if err != nil {
    fmt.Println("Error reading:", err.Error())
  }

  stringInput := strings.Trim(string(bufIn),"\x00")

  envelope := envelopeSSLServerHandshake([]byte(stringInput))

  sshServer.Write([]byte(envelope))
  return sshServer
}

func handleSshClientConnection(remoteAddress string,client net.Conn){
  sshServer := handleSshHandshakeClient(remoteAddress,client)
  Pipe(sshServer,client)
}

func handleSshServerConnection(remoteAddress string,client net.Conn){
  sshServer := handleSshHandshakeServer(remoteAddress,client)
  Pipe(sshServer,client)
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

func serveClient(localService string,remoteAddress string){

  ln, _ := net.Listen("tcp",localService)

  defer ln.Close()

  fmt.Println("Listening on :" + localService)

  for{
    conn, _ := ln.Accept()
    fmt.Printf("New connection established from '%v'\n", conn.RemoteAddr())
    go handleSshClientConnection(remoteAddress,conn)
  }
}

func serveServer(localService string,remoteSSHServer string){
  ln, _ := net.Listen("tcp",localService)
  defer ln.Close()

  fmt.Println("Listening on: "+localService)

  for{
    conn, _ := ln.Accept()
    fmt.Printf("New connection established from '%v'\n", conn.RemoteAddr())
    go handleSshServerConnection(remoteSSHServer,conn)
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
  cli.StringFlag{
			Name:   "from, f",
			Value:  "127.0.0.1:10000",
			EnvVar: "SSH_FROM",
			Usage:  "source HOST:PORT",
		},
		cli.StringFlag{
			Name:   "to, t",
			EnvVar: "SSH_TO",
			Usage:  "destination HOST:PORT",
		},
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
      serveServer(c.String("from"),c.String("to"))
    }else{
      serveClient(c.String("from"),c.String("to"))
    }
    return nil
  }

  app.Run(os.Args)

}
