package main

import (
	"crypto/sha256"
	"flag"
	"fmt"

	"github.com/aungmawjj/icmpnet"
	"github.com/aungmawjj/icmpnet/broker"
)

func main() {
	var password string
	flag.StringVar(&password, "pw", "password", "password")
	flag.Parse()

	sum := sha256.Sum256([]byte(password))
	aesKey := sum[:]

	server, err := icmpnet.NewServer(aesKey)
	check(err)
	b := broker.New()
	fmt.Println("Message broker started!")
	err = b.Serve(server)
	check(err)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
