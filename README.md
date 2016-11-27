# SSH to Http Packet wrapper

## Introduction

This is a tool to overcome the [deep packet inspection](https://en.wikipedia.org/wiki/Deep_packet_inspection) that is used against [secure shell](https://en.wikipedia.org/wiki/Secure_Shell) protocol. Deep packet inspection is a broad concept that involves use of several techniques with the same idea in common. It consists in deeply analyzing the network packets and applying rules and/or data mining over these same techniques. The use of deep packet inspection is morally questionable and poses a fundamental problem to the transparent use of Internet services. This is a simple tool that aims to avoid the filtering of SSH packets over a network that is being actively monitoring and droping this kind of packets.

## How do I know if the network is being DPI

Typically there are two different ways of blocking the use of a service in the network. The first consists in dropping all tcp packets from all the ports but a few. With this kind of blocking a simple *telnet host port* would end up in a refused or not allowed connection. The second one is a little more sneaky and does allow you to connect any port, or at least don't explicitly blocks you, instead it keeps analyzing the patterns inside the packets and when some pattern that is blacklisted like ssh or smtp handshake messages then it will drop following packets for that TCP connection. The fundamental difference is the first don't allow you even to establish a tcp connection while the second simply start dropping the following packets after the pattern is found and matched with an internal blacklist. So if you can connect to a <host:port> and suddently the traffic just stops to flow that is a strong indicator that your network is being actively monitored. If you wanna be sure about that you can simply change the protocol over that port if, for instance, you have control over the server that is hosting the service in that port. As an example you can just change ssh port with the HTTP port and retry the connections. What most certainly will happen is that the strange behavior has now swapped ports this kind of dynamic blocking is only possible because the packets are being deeply monitored and changed/drop depending in a set of rules defined by whom controls the network topology.


## How to use


### Installation

First you got to clone the project into your working space

    git clone git@github.com:Balhau/gossh2http.git

The next step we need to do is configure the GOROOT environment path

    export GOPATH=$HOME/<working_folder>/gossh2http


After you need to checkout some dependencies

    go get github.com/fatih/color
    go get github.com/urfave/cli

Then you need to go into the src folder and type

    go build ssh2http.go


### Running the executable

To check the command line documentation you can run

    ./ssh2http help

and get an output like the following. *Note that this is yet in development and more changes are to come*

    NAME:
     ssh2http - Ssh to http packet wrapping

    USAGE:
       ssh2http --from <local_ssh2http_ip>:<port --to <remote_ssh2http_tunnel>:<port>

    VERSION:
       1.0.0

    AUTHOR:
       Balhau <balhau@balhau.net>

    COMMANDS:
         help, h  Shows a list of commands or help for one command

    GLOBAL OPTIONS:
       --from value, -f value  source HOST:PORT (default: "127.0.0.1:10000") [$FWD_FROM]
       --to value, -t value    destination HOST:PORT [$FWD_TO]
       --serve, -s             list local addresses
       --help, -h              show help
       --version, -v           print the version

    COPYRIGHT:
       MIT License

To use this program you need to start the executable in two different points.

The idea behind this is explained in the following diagram



    |------------|    |---------------|  wrappedPackets |---------------|     |---------|
    | sshClient  | -->| wrapperClient | --------------->| wrapperServer | --->|sshServer|
    |------------|    |---------------|                 |---------------|     |---------|


So for this you need to start the wrapperServer in a machine outside the monitored network, and
a wrapperClient in your sshClient machine, the steps are the following


      sudo ./ssh2http -s -f localhost:10000 -t sshserver.com:22 --> In the server machine
      ./ssh2http -f localhost:10000 -t sshserver.com:10000
      ssh login@localhost:10000

As an running example you can check [here](https://www.youtube.com/watch?v=OK7DjsOyMp8) for a demo


# Notes

This tool was inspired in a very nice tool developed from a friend. [FWD](https://github.com/kintoandar/fwd)
Thanks @kintoandar for that
