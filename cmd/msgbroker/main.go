package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/aungmawjj/icmpnet"
	"github.com/aungmawjj/icmpnet/broker"
)

func main() {
	var password string
	flag.StringVar(&password, "pw", "password", "password")
	flag.Parse()

	sum := sha256.Sum256([]byte(password))
	aesKey := sum[:]

	rand.Seed(time.Now().UnixNano())

	ln, err := icmpnet.Listen(aesKey)
	check(err)

	welcome := fmt.Sprintf("Message Broker [ icmpnet ] %s\n", icmpnet.Version)
	b := broker.New(welcome)
	fmt.Println(welcome)

	err = b.Serve(ln)
	check(err)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
