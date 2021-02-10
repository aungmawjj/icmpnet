package rpc

import (
	"log"
	"net"
	"net/rpc"
)

// Service Methods
const (
	MethodInfo string = "info"
	MethodFile string = "files"
)

// Server type
type Server struct {
	rpcServer *rpc.Server
}

// NewServer creates a new Server
func NewServer(versionInfo, dirPath string) *Server {
	s := &Server{
		rpcServer: rpc.NewServer(),
	}
	s.rpcServer.RegisterName(MethodInfo, NewInfoAPI(versionInfo))
	s.rpcServer.RegisterName(MethodFile, NewFileAPI(dirPath))
	return s
}

// Serve handle rpc requests connections from listeners
func (s *Server) Serve(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go s.ServeConn(conn)
	}
}

// ServeConn serves a connection.
// It blocks until the connection is closed.
func (s *Server) ServeConn(conn net.Conn) {
	log.Printf("Connected: %s\n", conn.RemoteAddr())
	s.rpcServer.ServeConn(conn)
	log.Printf("Disconnected: %s\n", conn.RemoteAddr())
}
