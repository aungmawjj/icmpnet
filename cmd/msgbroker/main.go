package main

import (
	"crypto/sha256"
	"flag"

	"github.com/aungmawjj/icmpnet"
)

func main() {
	var password string
	flag.StringVar(&password, "pw", "password", "password")
	flag.Parse()

	sum := sha256.Sum256([]byte(password))
	aesKey := sum[:]

	server, err := icmpnet.NewServer(aesKey)
	check(err)
	b := icmpnet.NewBroker()
	b.Serve(server)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
