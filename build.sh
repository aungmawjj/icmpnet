#!/bin/sh

go build -o ./bin/ ./cmd/msgclient
go build -o ./bin/msgbroker ./cmd/msgbroker

go build -o ./bin/ ./cmd/fileclient
go build -o ./bin/ ./cmd/fileserver