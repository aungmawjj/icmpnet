package icmpnet

import (
	"fmt"
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
	id     uint16
	readCh chan *icmp.Message
	host   host
}

func newICMPConn(h host, id int, addr net.Addr) *icmpConn {
	ic := &icmpConn{
		bufferConn: *newBufferConn(h.localAddr(), addr),
		id:         uint16(id),
		readCh:     make(chan *icmp.Message, 100),
		host:       h,
	}
	return ic
}

func newICMPServerConn(h host, id int, addr net.Addr) *icmpConn {
	ic := newICMPConn(h, id, addr)
	go ic.serverLoop()
	return ic
}

func newICMPClientConn(h host, id int, addr net.Addr) *icmpConn {
	ic := newICMPConn(h, id, addr)
	go ic.clientLoop()
	return ic
}

func (ic *icmpConn) serverLoop() {
	defer func() {
		ic.Close()
		ic.host.onConnClose(ic)
	}()

	var (
		n    int
		body *icmp.Echo
		ok   bool
		err  error
	)

	prevSeq := 0
	buf := make([]byte, 4096)

	for {
		select {
		case msg := <-ic.readCh:
			isNewMsg := false
			body, ok = msg.Body.(*icmp.Echo)
			if ok {
				if body.Seq == prevSeq+1 || body.Seq == 1 {
					isNewMsg = true
					prevSeq = body.Seq
					ic.writeInBuf(body.Data)
				}
			}

			if isNewMsg {
				n, err = ic.readOutBuf(buf)
				if err != nil {
					return
				}
			}

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

	var (
		n   int
		err error
	)

	seq := 1
	prevSeq := 0
	buf := make([]byte, 4096)

	for {
		if seq == prevSeq+1 {
			prevSeq = seq
			n, err = ic.readOutBuf(buf)
			if err != nil {
				return
			}
		}
		body := &icmp.Echo{
			ID:   int(ic.id),
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
				if seq == body.Seq {
					ic.writeInBuf(body.Data)
					seq++
				}
			}
		case <-time.After(2 * time.Second):
			// fmt.Println("not received reply")
		}
	}
}

func (ic *icmpConn) String() string {
	return icmpConnKey(ic.remoteAddr, ic.ID())
}

func (ic *icmpConn) ID() int {
	return int(ic.id)
}

func icmpConnKey(remoteAddr net.Addr, id int) string {
	return fmt.Sprintf("%s-%d", remoteAddr.String(), id)
}
