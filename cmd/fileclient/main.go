package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/aungmawjj/icmpnet"
	"github.com/aungmawjj/icmpnet/rpc"
)

type endProgressFunc func()

func showProgress() endProgressFunc {
	end := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-end:
				return
			case <-time.After(300 * time.Millisecond):
				fmt.Print(".")
			}
		}
	}()
	return func() { close(end) }
}

func uploadFile(client *rpc.Client) {
	var filename string
	fmt.Print("Enter file path >>  ")
	fmt.Scanln(&filename)

	endP := showProgress()
	err := client.FileUpload(filename)
	endP()
	check(err)

	fmt.Print("\n\n")
	fmt.Println("File upload successful!")
}

func downloadFile(client *rpc.Client) {
	var filename string
	fmt.Print("Enter file name >>  ")
	fmt.Scanln(&filename)

	dir, err := os.Getwd()
	check(err)

	endP := showProgress()
	err = client.FileDownload(filename, dir)
	endP()
	check(err)

	fmt.Print("\n\n")
	fmt.Println("File download successful!")
}

func main() {
	var (
		serverIP      string
		password      string
		inputServerIP string
		inputPassword string
		mode          int
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

	sum := sha256.Sum256([]byte(password))
	aesKey := sum[:]

	addr, err := net.ResolveIPAddr("ip4", serverIP)
	check(err)

	fmt.Printf("Connecting: %s ...\n", addr)
	conn, err := icmpnet.Connect(addr, aesKey)
	check(err)
	fmt.Print("Connected! Uploading File...\n\n")

	rpcClient := rpc.NewClient(conn)

	vInfo, err := rpcClient.InfoVersion()
	if err == nil {
		fmt.Println(vInfo)
	}

	fmt.Print("Select mode Upload = 1, Download = 2 >>  ")
	fmt.Scanln(&mode)
	switch mode {
	case 1:
		uploadFile(rpcClient)
	case 2:
		downloadFile(rpcClient)
	default:
		fmt.Println("Unknown mode")
	}

}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
