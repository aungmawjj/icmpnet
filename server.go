package icmpnet

import (
	"crypto/aes"
	"fmt"
	"log"
	"net"
	"sync"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// Server type
type Server struct {
	aesKey []byte
	pconn  *icmp.PacketConn
	cpMtx  sync.RWMutex

	sessions map[string]*session
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
		sessions: make(map[string]*session),
		connCh:   make(chan net.Conn, 100),
	}
	go s.mainLoop()
	return s, nil
}

func (s *Server) mainLoop() {
	buf := make([]byte, 32768)
	for {
		n, addr, err := s.pconn.ReadFrom(buf)
		check(err)
		msg, err := icmp.ParseMessage(1, buf[:n])
		if err != nil {
			continue
		}
		if msg.Type != ipv4.ICMPTypeEcho {
			continue
		}
		sess := s.loadSession(addr.String())
		if sess == nil {
			sess = &session{
				bufferConn: newBufferConn(s.pconn.LocalAddr(), addr),
				readCh:     make(chan *icmp.Message, 100),
				sendMsg:    s.sendTo(addr),
				onClose:    s.onSessionClose(addr.String()),
			}
			if s.aesKey == nil {
				sess.conn = sess.bufferConn
			} else {
				sess.conn, _ = newSecureConn(sess.bufferConn, s.aesKey)
			}
			log.Printf("New Connection: %s\n", addr)
			go sess.mainloop()
			s.storeSession(addr.String(), sess)
			s.emitNewConn(sess.conn)
		}
		sess.readCh <- msg
	}
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
		conns := s.AllConns()
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

// AllConns ...
func (s *Server) AllConns() []net.Conn {
	s.cpMtx.RLock()
	defer s.cpMtx.RUnlock()
	ret := make([]net.Conn, 0, len(s.sessions))
	for _, sess := range s.sessions {
		ret = append(ret, sess.conn)
	}
	return ret
}

func (s *Server) loadSession(key string) *session {
	s.cpMtx.RLock()
	defer s.cpMtx.RUnlock()
	return s.sessions[key]
}

func (s *Server) storeSession(key string, conn *session) {
	s.cpMtx.Lock()
	defer s.cpMtx.Unlock()
	s.sessions[key] = conn
}

func (s *Server) onSessionClose(key string) func() {
	return func() {
		log.Printf("Closed Connection: %s\n", key)
		s.deleteSession(key)
	}
}

func (s *Server) deleteSession(key string) {
	s.cpMtx.Lock()
	defer s.cpMtx.Unlock()
	delete(s.sessions, key)
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
