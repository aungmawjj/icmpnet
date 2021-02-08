package rpc

import (
	"log"
	"net"
	"net/rpc"
)

// Service Methods
const (
	MethodFile string = "files"
)

// Server type
type Server struct {
	rpcServer *rpc.Server
}

// NewServer creates a new Server
func NewServer(dirPath string) *Server {
	s := &Server{
		rpcServer: rpc.NewServer(),
	}
	s.rpcServer.RegisterName(MethodFile, NewFileAPI(dirPath))
	return s
}

// Serve ...
func (s *Server) Serve(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		log.Printf("New Connection: %s\n", conn.RemoteAddr())
		go s.rpcServer.ServeConn(conn)
	}
}
