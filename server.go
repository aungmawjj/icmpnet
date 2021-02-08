package icmpnet

import (
	"crypto/aes"
	"fmt"
	"net"
	"sync"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// Server type
type Server struct {
	aesKey []byte
	pconn  *icmp.PacketConn

	connPool map[string]*icmpConn
	cpMtx    sync.RWMutex
	connCh   chan net.Conn

	closed chan struct{}
}

// NewServer create a new Server.
// When aesKey is nil, encryption is disabled.
func NewServer(aesKey []byte) (*Server, error) {
	// verify aesKey
	if aesKey != nil {
		if _, err := aes.NewCipher(aesKey); err != nil {
			return nil, err
		}
	}

	pconn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, err
	}
	s := &Server{
		aesKey:   aesKey,
		pconn:    pconn,
		connPool: make(map[string]*icmpConn),
		connCh:   make(chan net.Conn, 100),
	}
	go s.mainLoop()
	return s, nil
}

func (s *Server) mainLoop() {
	buf := make([]byte, 5000)
	for {
		n, addr, err := s.pconn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		msg, err := icmp.ParseMessage(1, buf[:n])
		if err != nil {
			continue
		}
		if msg.Type != ipv4.ICMPTypeEcho {
			continue
		}
		conn := s.loadConn(addr.String())
		if conn == nil {
			conn = s.newICMPconn(addr)
			s.storeConn(addr.String(), conn)

			if s.aesKey == nil {
				s.emitNewConn(conn)
			} else {
				sconn, _ := newSecureConn(conn, s.aesKey)
				s.emitNewConn(sconn)
			}
		}
		conn.readCh <- msg
	}
}

func (s *Server) newICMPconn(addr net.Addr) *icmpConn {
	conn := &icmpConn{
		bufferConn: *newBufferConn(s.pconn.LocalAddr(), addr),
		readCh:     make(chan *icmp.Message, 100),
		sendMsg:    s.sendTo(addr),
		onClose:    s.onConnClose(addr.String()),
		onConnect:  func() {},
	}
	go conn.serverLoop()
	return conn
}

func (s *Server) emitNewConn(conn net.Conn) {
	select {
	case s.connCh <- conn:
	default:
	}
}

// Accept implements net.Listener
func (s *Server) Accept() (net.Conn, error) {
	select {
	case conn := <-s.connCh:
		return conn, nil
	case <-s.closed:
		return nil, fmt.Errorf("closed")
	}
}

// Close implements net.Listener
func (s *Server) Close() error {
	select {
	case <-s.closed:
		return fmt.Errorf("closed")
	default:
		close(s.closed)
		conns := s.allConns()
		for _, conn := range conns {
			conn.Close()
		}
		return nil
	}
}

// Addr implements net.Listener
func (s *Server) Addr() net.Addr {
	return s.pconn.LocalAddr()
}

func (s *Server) allConns() []*icmpConn {
	s.cpMtx.RLock()
	defer s.cpMtx.RUnlock()
	ret := make([]*icmpConn, 0, len(s.connPool))
	for _, conn := range s.connPool {
		ret = append(ret, conn)
	}
	return ret
}

func (s *Server) loadConn(key string) *icmpConn {
	s.cpMtx.RLock()
	defer s.cpMtx.RUnlock()
	return s.connPool[key]
}

func (s *Server) storeConn(key string, conn *icmpConn) {
	s.cpMtx.Lock()
	defer s.cpMtx.Unlock()
	s.connPool[key] = conn
}

func (s *Server) onConnClose(key string) func() {
	return func() {
		s.deleteConn(key)
	}
}

func (s *Server) deleteConn(key string) {
	s.cpMtx.Lock()
	defer s.cpMtx.Unlock()
	delete(s.connPool, key)
}

func (s *Server) sendTo(addr net.Addr) func(msg *icmp.Message) error {
	return func(msg *icmp.Message) error {
		b, err := msg.Marshal(nil)
		if err != nil {
			return err
		}
		_, err = s.pconn.WriteTo(b, addr)
		return err
	}
}
