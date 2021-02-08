package icmpnet

import (
	"math/rand"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type icmpConn struct {
	bufferConn
	readCh    chan *icmp.Message
	sendMsg   func(msg *icmp.Message) error
	onClose   func()
	onConnect func()
}

func (ic *icmpConn) serverLoop() {
	defer func() {
		ic.Close()
		ic.onClose()
	}()

	buf := make([]byte, 32768)
	for {
		select {
		case msg := <-ic.readCh:
			var (
				body *icmp.Echo
				ok   bool
			)
			if body, ok = msg.Body.(*icmp.Echo); ok {
				ic.writeInBuf(body.Data)
			}
			n, err := ic.readOutBuf(buf)
			if err != nil {
				return
			}

			if n == 0 && len(body.Data) == 0 {
				time.Sleep(300 * time.Millisecond)
			}
			body.Data = buf[:n]
			msg.Type = ipv4.ICMPTypeEchoReply
			err = ic.sendMsg(msg)
			if err != nil {
				return
			}

		case <-time.After(2 * time.Second):
			// packet lost can occur
			return
		}
	}
}

func (ic *icmpConn) clientLoop() {
	defer func() {
		ic.Close()
		ic.onClose()
	}()

	id := rand.Int()
	seq := 1

	buf := make([]byte, 32768)
	for {
		n, err := ic.readOutBuf(buf)
		if err != nil {
			return
		}

		body := &icmp.Echo{
			ID:   id,
			Seq:  seq,
			Data: buf[:n],
		}
		msg := &icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: body,
		}
		err = ic.sendMsg(msg)
		if err != nil {
			return
		}

		select {
		case msg := <-ic.readCh:
			if seq == 1 {
				ic.onConnect()
			}
			if body, ok := msg.Body.(*icmp.Echo); ok {
				ic.writeInBuf(body.Data)
			}
		case <-time.After(2 * time.Second):
			// packet lost can occur
		}
		seq++
	}
}
