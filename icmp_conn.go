package icmpnet

import (
	"math/rand"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type host interface {
	sendMsg(msg *icmp.Message, addr net.Addr) error
	onConnClose(conn *icmpConn)
	localAddr() net.Addr
}

type icmpConn struct {
	bufferConn
	readCh chan *icmp.Message
	host   host
}

func newICMPconn(h host, addr net.Addr, server bool) *icmpConn {
	ic := &icmpConn{
		bufferConn: *newBufferConn(h.localAddr(), addr),
		readCh:     make(chan *icmp.Message, 100),
		host:       h,
	}
	if server {
		go ic.serverLoop()
	} else {
		go ic.clientLoop()
	}
	return ic
}

func (ic *icmpConn) serverLoop() {
	defer func() {
		ic.Close()
		ic.host.onConnClose(ic)
	}()

	prevSeq := 0

	buf := make([]byte, 4096)
	for {
		select {
		case msg := <-ic.readCh:
			var (
				n    int
				body *icmp.Echo
				ok   bool
				err  error
			)
			body, ok = msg.Body.(*icmp.Echo)
			if ok {
				if body.Seq > prevSeq {
					ic.writeInBuf(body.Data)
				}
			}

			if body.Seq > prevSeq {
				n, err = ic.readOutBuf(buf)
				if err != nil {
					return
				}
			}

			prevSeq = body.Seq

			if n == 0 && len(body.Data) == 0 {
				time.Sleep(300 * time.Millisecond)
			}

			body.Data = buf[:n]
			msg.Type = ipv4.ICMPTypeEchoReply
			err = ic.host.sendMsg(msg, ic.remoteAddr)
			if err != nil {
				return
			}

		case <-time.After(5 * time.Second):
			// fmt.Println("packet may lost")
			return
		}
	}
}

func (ic *icmpConn) clientLoop() {
	defer func() {
		ic.Close()
		ic.host.onConnClose(ic)
	}()

	id := rand.Int()
	seq := 1
	prevSeq := 0

	buf := make([]byte, 4096)
	for {
		var (
			n   int
			err error
		)
		if seq > prevSeq {
			n, err = ic.readOutBuf(buf)
			if err != nil {
				return
			}
			prevSeq = seq
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
		err = ic.host.sendMsg(msg, ic.remoteAddr)
		if err != nil {
			return
		}

		select {
		case msg := <-ic.readCh:
			if body, ok := msg.Body.(*icmp.Echo); ok {
				ic.writeInBuf(body.Data)
			}
			seq++
		case <-time.After(1 * time.Second):
			// fmt.Println("packet may lost")
		}
	}
}
