package main

import (
	"bufio"
	"crypto/sha256"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/aungmawjj/icmpnet"
	"github.com/aungmawjj/icmpnet/rpc"
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
		password string
		filepath string
	)
	flag.StringVar(&serverIP, "server", "13.212.27.85", "server ip address")
	flag.StringVar(&password, "pw", "password", "password")
	flag.StringVar(&filepath, "path", "", "file path")
	flag.Parse()

	sum := sha256.Sum256([]byte(password))
	aesKey := sum[:]

	addr, err := net.ResolveIPAddr("ip4", serverIP)
	check(err)

	fmt.Printf("Connecting: %s ...\n", addr)
	conn, err := icmpnet.Connect(addr, aesKey)
	check(err)
	fmt.Print("Connected! Uploading File...\n\n")

	endProgress := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-endProgress:
				return
			case <-time.After(500 * time.Millisecond):
				fmt.Print(".")
			}
		}
	}()

	rpcClient := rpc.NewClient(conn)
	err = rpcClient.FileUpload(filepath)
	endProgress <- struct{}{}
	fmt.Print("\n\n")
	check(err)
	fmt.Println("File upload successful!")
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
