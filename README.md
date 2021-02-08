# Messaging over ICMP
Implementation of client/server bidirectional messaging over ICMP protocol using golang.

Messaging over ICMP is useful - when your network (wifi) gives you an IP address, but won't let you send TCP or UDP packets out to the rest of the internet, but allow you to ping any computer on the internet.

Features:
- AES encryption is used.
- Implements standard net.Listener and net.Conn interface to be able to extend for high level protocols such as http, rpc.

Implemented Use-case applications:
- simple message broker and client. (each message is a single line)
- rpc server and client

## Build
```sh
./build.sh
```

## Run

### Message Broker Application

Broker
```sh
// stop auto reply ping messages for linux
echo 1 | sudo dd of=/proc/sys/net/ipv4/icmp_echo_ignore_all
sudo ./msgbroker -pw <password>
```

Client
```sh
sudo ./msgclient -server <serverIP> -name <Your Name> -pw <password>
```

### RPC Application

Coming soon ;)

## Using API

You can use Client and Server types to establish a baseline connection with aes encryption.

Please check existing applications in [cmd folder](cmd).

## License

This repository is under MIT License, as found in [LICENSE file](LICENSE).