package icmpnet

import (
	"net"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// Client type
type Client struct {
	pconn       *icmp.PacketConn
	conn        *icmpConn
	connectedCh chan struct{}
}

// NewClient creates a new Client
func NewClient() (*Client, error) {
	pconn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, err
	}
	return &Client{
		pconn:       pconn,
		connectedCh: make(chan struct{}, 1),
	}, nil
}

// Connect create a connection to server.
// When aesKey is nil, encryption is disabled.
func (c *Client) Connect(server net.Addr, aesKey []byte) (net.Conn, error) {
	c.conn = &icmpConn{
		bufferConn: *newBufferConn(c.pconn.LocalAddr(), server),
		readCh:     make(chan *icmp.Message, 100),
		sendMsg:    c.sendMsg,
		onClose:    func() {},
		onConnect:  c.onConnect,
	}
	go c.conn.clientLoop()
	go c.mainLoop()

	<-c.connectedCh

	if aesKey == nil {
		return c.conn, nil
	}
	return newSecureConn(c.conn, aesKey)
}

func (c *Client) onConnect() {
	c.connectedCh <- struct{}{}
}

func (c *Client) mainLoop() {
	buf := make([]byte, 5000)
	for {
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
		if err == nil {
			c.conn.readCh <- msg
		}
	}
}

func (c *Client) sendMsg(msg *icmp.Message) error {
	b, err := msg.Marshal(nil)
	if err != nil {
		return err
	}
	_, err = c.pconn.WriteTo(b, c.conn.remoteAddr)
	return err
}
