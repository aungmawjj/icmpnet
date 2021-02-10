package icmpnet

import (
	"math/rand"
	"net"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type client struct {
	pconn       *icmp.PacketConn
	conn        *icmpConn
	closedCh    chan struct{}
	connectedCh chan struct{}
}

// Connect create a connection to server.
// If aesKey is nil, encryption is disabled.
func Connect(server net.Addr, aesKey []byte) (net.Conn, error) {
	pconn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, err
	}
	c := &client{
		pconn:       pconn,
		connectedCh: make(chan struct{}, 1),
	}
	go c.mainLoop()

	c.conn = newICMPClientConn(c, rand.Int(), server)
	<-c.connectedCh

	if aesKey == nil {
		return c.conn, nil
	}
	return newSecureConn(c.conn, aesKey)
}

func (c *client) mainLoop() {
	buf := make([]byte, 5000)
	connected := false
	for {
		select {
		case <-c.closedCh:
			return
		default:
			n, addr, err := c.pconn.ReadFrom(buf)
			if err != nil {
				panic(err)
			}
			msg, err := icmp.ParseMessage(1, buf[:n])
			if err != nil {
				continue
			}
			if msg.Type != ipv4.ICMPTypeEchoReply {
				continue
			}
			if addr.String() != c.conn.RemoteAddr().String() {
				continue
			}
			msg, err = icmp.ParseMessage(1, buf[:n])
			if err != nil {
				continue
			}
			if body, ok := msg.Body.(*icmp.Echo); ok {
				if body.ID != c.conn.ID() {
					continue
				}
				if !connected {
					connected = true
					c.connectedCh <- struct{}{}
				}
				c.conn.readCh <- msg
			}
		}
	}
}

func (c *client) localAddr() net.Addr {
	return c.pconn.LocalAddr()
}

func (c *client) onConnClose(conn *icmpConn) {
	select {
	case <-c.closedCh:
	default:
		close(c.closedCh)
	}
}

func (c *client) sendMsg(msg *icmp.Message, addr net.Addr) error {
	b, err := msg.Marshal(nil)
	if err != nil {
		return err
	}
	_, err = c.pconn.WriteTo(b, addr)
	return err
}
