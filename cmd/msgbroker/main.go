package main

import (
	"bufio"
	"crypto/sha256"
	"flag"
	"log"
	"net"

	"github.com/aungmawjj/icmpnet"
)

type broker struct {
	server *icmpnet.Server
}

func (b *broker) Serve() {
	for {
		conn, err := b.server.Accept()
		check(err)
		go b.serveConn(conn)
	}
}

func (b *broker) serveConn(conn net.Conn) {
	r := bufio.NewReader(conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			break
		}
		log.Printf("%s >> %s", conn.RemoteAddr(), msg)
		b.broadcast(msg)
	}
}

func (b *broker) broadcast(msg string) {
	conns := b.server.AllConns()
	for _, conn := range conns {
		conn.Write([]byte(msg))
	}
}

func main() {
	var password string
	flag.StringVar(&password, "pw", "password", "password")
	flag.Parse()

	sum := sha256.Sum256([]byte(password))
	aesKey := sum[:]

	server, err := icmpnet.NewServer(aesKey)
	check(err)
	b := &broker{server}
	b.Serve()
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
