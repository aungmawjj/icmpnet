package icmpnet

import (
	"crypto/aes"
	"fmt"
	"net"
	"sync"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type server struct {
	aesKey []byte
	pconn  *icmp.PacketConn

	connPool map[string]*icmpConn
	cpMtx    sync.RWMutex
	connCh   chan net.Conn

	closedCh chan struct{}
}

// Listen creates a new icmp listener (server).
// When aesKey is nil, encryption is disabled.
func Listen(aesKey []byte) (net.Listener, error) {
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
	s := &server{
		aesKey:   aesKey,
		pconn:    pconn,
		connPool: make(map[string]*icmpConn),
		connCh:   make(chan net.Conn, 100),
	}
	go s.mainLoop()
	return s, nil
}

// Accept implements net.Listener
func (s *server) Accept() (net.Conn, error) {
	select {
	case conn := <-s.connCh:
		return conn, nil
	case <-s.closedCh:
		return nil, fmt.Errorf("closedCh")
	}
}

// Close implements net.Listener
func (s *server) Close() error {
	select {
	case <-s.closedCh:
		return fmt.Errorf("closedCh")
	default:
		close(s.closedCh)
		conns := s.allConns()
		for _, conn := range conns {
			conn.Close()
		}
		return nil
	}
}

// Addr implements net.Listener
func (s *server) Addr() net.Addr {
	return s.pconn.LocalAddr()
}

func (s *server) mainLoop() {
	buf := make([]byte, 5000)
	for {
		select {
		case <-s.closedCh:
			return
		default:
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
}

func (s *server) newICMPconn(addr net.Addr) *icmpConn {
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

func (s *server) emitNewConn(conn net.Conn) {
	select {
	case s.connCh <- conn:
	default:
	}
}

func (s *server) allConns() []*icmpConn {
	s.cpMtx.RLock()
	defer s.cpMtx.RUnlock()
	ret := make([]*icmpConn, 0, len(s.connPool))
	for _, conn := range s.connPool {
		ret = append(ret, conn)
	}
	return ret
}

func (s *server) loadConn(key string) *icmpConn {
	s.cpMtx.RLock()
	defer s.cpMtx.RUnlock()
	return s.connPool[key]
}

func (s *server) storeConn(key string, conn *icmpConn) {
	s.cpMtx.Lock()
	defer s.cpMtx.Unlock()
	s.connPool[key] = conn
}

func (s *server) onConnClose(key string) func() {
	return func() {
		s.deleteConn(key)
	}
}

func (s *server) deleteConn(key string) {
	s.cpMtx.Lock()
	defer s.cpMtx.Unlock()
	delete(s.connPool, key)
}

func (s *server) sendTo(addr net.Addr) func(msg *icmp.Message) error {
	return func(msg *icmp.Message) error {
		b, err := msg.Marshal(nil)
		if err != nil {
			return err
		}
		_, err = s.pconn.WriteTo(b, addr)
		return err
	}
}
