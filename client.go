package icmpnet

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// Client type
type Client struct {
	id    int
	seq   int
	pconn *icmp.PacketConn
	conn  *bufferConn
	buf   []byte
}

// NewClient creates a new Client
func NewClient() (*Client, error) {
	pconn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, err
	}
	return &Client{
		id:    rand.Int(),
		seq:   1,
		pconn: pconn,
		buf:   make([]byte, 32768),
	}, nil
}

// Connect create a connection to server.
// When aesKey is nil, encryption is disabled.
func (c *Client) Connect(server net.Addr, aesKey []byte) (net.Conn, error) {
	c.conn = newBufferConn(c.pconn.LocalAddr(), server)
	go c.mainloop()
	if aesKey == nil {
		return c.conn, nil
	}
	return newSecureConn(c.conn, aesKey)
}

func (c *Client) mainloop() {
	connected := false
	readCh := make(chan *icmp.Message, 5)
	fmt.Println("Connecting...")
	for {
		n, err := c.conn.readOutBuf(c.buf)
		check(err)

		body := &icmp.Echo{
			ID:   c.id,
			Seq:  c.seq,
			Data: c.buf[:n],
		}
		msg := &icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: body,
		}
		err = c.sendMsg(msg)
		check(err)

		c.readMsgCh(readCh)
		select {
		case msg := <-readCh:
			if connected == false {
				connected = true
				fmt.Print("Connected!\n\n")
			}
			if body, ok := msg.Body.(*icmp.Echo); ok {
				c.conn.writeInBuf(body.Data)
			}
		case <-time.After(2 * time.Second):
		}
		c.seq++
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

func (c *Client) readMsgCh(ch chan<- *icmp.Message) {
	go func() {
		msg, err := c.readMsg()
		check(err)
		ch <- msg
	}()
}

func (c *Client) readMsg() (*icmp.Message, error) {
	for {
		n, addr, err := c.pconn.ReadFrom(c.buf)
		if err != nil {
			return nil, err
		}
		if addr.String() != c.conn.RemoteAddr().String() {
			continue
		}
		return icmp.ParseMessage(1, c.buf[:n])
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
