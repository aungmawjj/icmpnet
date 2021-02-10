package main

import (
	"bufio"
	"crypto/sha256"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

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
		serverIP      string
		password      string
		inputServerIP string
		inputPassword string
		username      string
	)
	flag.StringVar(&serverIP, "server", "13.212.27.85", "server ip address")
	flag.StringVar(&password, "pw", "password", "password")
	flag.Parse()

	fmt.Printf("Enter server ip [default = %s] >>  ", serverIP)
	fmt.Scanln(&inputServerIP)
	if inputServerIP != "" {
		serverIP = inputServerIP
	}

	fmt.Printf("Enter password ip [default = %s] >>  ", password)
	fmt.Scanln(&inputPassword)
	if inputPassword != "" {
		password = inputPassword
	}

	fmt.Print("Enter username >>  ")
	fmt.Scanln(&username)
	if username == "" {
		fmt.Println("Must provide user name")
		return
	}

	sum := sha256.Sum256([]byte(password))
	aesKey := sum[:]

	addr, err := net.ResolveIPAddr("ip4", serverIP)
	check(err)

	rand.Seed(time.Now().UnixNano()) // to generate random client id

	fmt.Printf("Connecting: %s ...\n", addr)
	conn, err := icmpnet.Connect(addr, aesKey)
	check(err)
	fmt.Print("Connected!\n\n")

	go printIncoming(conn)
	go sendMessages(conn, username)
	select {}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
