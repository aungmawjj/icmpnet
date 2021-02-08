package main

import (
	"crypto/sha256"
	"flag"
	"fmt"

	"github.com/aungmawjj/icmpnet"
	"github.com/aungmawjj/icmpnet/rpc"
)

func main() {
	var (
		password string
		dirPath  string
	)
	flag.StringVar(&password, "pw", "password", "password")
	flag.StringVar(&dirPath, "dir", "uploaded_files", "directory for uploaded files")
	flag.Parse()

	sum := sha256.Sum256([]byte(password))
	aesKey := sum[:]

	icmpServer, err := icmpnet.NewServer(aesKey)
	check(err)

	rpcServer := rpc.NewServer(dirPath)
	fmt.Println("File server started!")
	err = rpcServer.Serve(icmpServer)
	check(err)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
