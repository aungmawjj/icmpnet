# Messaging over ICMP
[![GoDoc](https://godoc.org/github.com/aungmawjj/icmpnet?status.svg)](https://pkg.go.dev/github.com/aungmawjj/icmpnet#section-documentation)

Implementation of client/server bidirectional messaging over ICMP protocol using golang.

Messaging over ICMP is useful - when your network (wifi) gives you an IP address, but won't let you send TCP or UDP packets out to the rest of the internet, but allow you to ping any computer on the internet.

Features:
- AES encryption is used.
- Implements standard net.Listener and net.Conn interface to be able to extend for high level protocols such as http, rpc.

Implemented Use-case applications:
- Message broker and client (each message is a single line)
- File upload (used rpc to upload file and respond status)

## Build
```sh
./build.sh
```

## Run

### Message Broker Application

Broker
```sh
# stop auto reply ping messages for linux
echo 1 | sudo dd of=/proc/sys/net/ipv4/icmp_echo_ignore_all
sudo ./bin/msgbroker -pw <password>
```

Client
```sh
sudo ./bin/msgclient -server <serverIP> -pw <password> -name <Your Name>
```

### File Upload Application

Server
```sh
# stop auto reply ping messages for linux
echo 1 | sudo dd of=/proc/sys/net/ipv4/icmp_echo_ignore_all
sudo ./bin/fileserver -pw <password> -dir <file_directory>
```

Client
```sh
sudo ./bin/fileclient -server <serverIP> -pw <password> -path <filepath>
```


## Using the API
[Reference](https://pkg.go.dev/github.com/aungmawjj/icmpnet#section-documentation)

Listen connections at server
```go
listener, err := icmpnet.Listen(aesKey)
```

Connect to server
```go
addr, _ := net.ResolveIPAddr("ip4", "server_IP")
conn, err := icmpnet.Connect(serverAddr, aesKey)
```

Please check sample applications in [cmd folder](cmd).

## License

This repository is under MIT License, as found in [LICENSE file](LICENSE).
