#!/bin/sh

env GOOS=windows go build -o ./bin/msgclient-windows ./cmd/msgclient
env GOOS=linux go build -o ./bin/msgclient-linux ./cmd/msgclient
env GOOS=darwin go build -o ./bin/msgclient-mac ./cmd/msgclient

env GOOS=linux go build -o ./bin/msgbroker ./cmd/msgbroker