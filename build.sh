#!/bin/sh

env GOOS=windows go build -o ./bin/msgclient-windows ./cmd/msgclient
env GOOS=linux go build -o ./bin/msgclient-linux ./cmd/msgclient
env GOOS=darwin go build -o ./bin/msgclient-mac ./cmd/msgclient

env GOOS=linux go build -o ./bin/msgbroker ./cmd/msgbroker


env GOOS=windows go build -o ./bin/fileclient-windows ./cmd/fileclient
env GOOS=linux go build -o ./bin/fileclient-linux ./cmd/fileclient
env GOOS=darwin go build -o ./bin/fileclient-mac ./cmd/fileclient

env GOOS=linux go build -o ./bin/fileserver ./cmd/fileserver