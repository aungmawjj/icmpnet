package main

import (
	"bufio"
	"crypto/sha256"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/aungmawjj/icmpnet"
)

func printIncoming(conn net.Conn) {
	r := bufio.NewReader(conn)
	for {
		msg, err := r.ReadString('\n')
		check(err)
		fmt.Printf(">> %s\n", msg)
	}
}

func sendMessages(conn net.Conn, nickName string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		msg, err := reader.ReadString('\n')
		check(err)
		_, err = fmt.Fprintf(conn, "[ %s ] >>  %s", nickName, msg)
		check(err)
	}
}

func main() {
	var (
		serverIP string
		nickName string
		password string
	)
	flag.StringVar(&serverIP, "server", "13.212.27.85", "server ip address")
	flag.StringVar(&password, "pw", "password", "password")
	flag.StringVar(&nickName, "name", "", "nick name")
	flag.Parse()

	if nickName == "" {
		panic("must set a nick name")
	}

	sum := sha256.Sum256([]byte(password))
	aesKey := sum[:]

	addr, err := net.ResolveIPAddr("ip4", serverIP)
	check(err)

	fmt.Printf("Connecting: %s ...\n", addr)
	conn, err := icmpnet.Connect(addr, aesKey)
	check(err)
	fmt.Print("Connected! Enter messages.\n\n")

	go printIncoming(conn)
	go sendMessages(conn, nickName)
	select {}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
