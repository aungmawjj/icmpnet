package icmpnet

import (
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type session struct {
	bufferConn *bufferConn
	conn       net.Conn
	readCh     chan *icmp.Message
	sendMsg    func(msg *icmp.Message) error
	onClose    func()
}

func (s *session) mainloop() {
	defer func() {
		s.conn.Close()
		s.onClose()
	}()

	buf := make([]byte, 32768)
	for {
		select {
		case msg := <-s.readCh:
			var (
				body *icmp.Echo
				ok   bool
			)
			if body, ok = msg.Body.(*icmp.Echo); ok {
				s.bufferConn.writeInBuf(body.Data)
			}
			n, err := s.bufferConn.readOutBuf(buf)
			if err != nil {
				return
			}

			if n == 0 && len(body.Data) == 0 {
				time.Sleep(300 * time.Millisecond)
			}
			body.Data = buf[:n]
			msg.Type = ipv4.ICMPTypeEchoReply
			err = s.sendMsg(msg)
			if err != nil {
				return
			}

		case <-time.After(2 * time.Second):
			return
		}
	}
}
