package icmpnet

import (
	"bytes"
	"io"
	"net"
	"sync"
	"time"
)

type bufferConn struct {
	localAddr  net.Addr
	remoteAddr net.Addr

	inBuf  *bytes.Buffer
	outBuf *bytes.Buffer

	inMtx  sync.Mutex
	outMtx sync.Mutex

	inCh   chan struct{}
	closed chan struct{}
}

var _ net.Conn = (*bufferConn)(nil)

func newBufferConn(localAddr, remoteAddr net.Addr) *bufferConn {
	return &bufferConn{
		remoteAddr: remoteAddr,
		inBuf:      bytes.NewBuffer(nil),
		outBuf:     bytes.NewBuffer(nil),
		inCh:       make(chan struct{}),
		closed:     make(chan struct{}),
	}
}

func (c *bufferConn) Read(b []byte) (n int, err error) {
	select {
	case <-c.closed:
		return 0, io.EOF
	default:
		n, err = c.readInBuf(b)
		if err == io.EOF {
			select {
			case <-c.closed:
				return 0, io.EOF
			case <-c.inCh:
				return c.Read(b)
			}
		}
		return n, err
	}
}

func (c *bufferConn) readInBuf(b []byte) (int, error) {
	c.inMtx.Lock()
	defer c.inMtx.Unlock()
	return c.inBuf.Read(b)
}

func (c *bufferConn) writeInBuf(b []byte) (n int, err error) {
	c.inMtx.Lock()
	defer c.inMtx.Unlock()
	n, err = c.inBuf.Write(b)
	select {
	case c.inCh <- struct{}{}:
	default:
	}
	return n, err
}

func (c *bufferConn) Write(b []byte) (n int, err error) {
	select {
	case <-c.closed:
		return 0, io.ErrClosedPipe
	default:
		return c.writeOutBuf(b)
	}
}

func (c *bufferConn) writeOutBuf(b []byte) (n int, err error) {
	c.outMtx.Lock()
	defer c.outMtx.Unlock()
	n, err = c.outBuf.Write(b)
	return n, err
}

func (c *bufferConn) readOutBuf(b []byte) (n int, err error) {
	c.outMtx.Lock()
	defer c.outMtx.Unlock()
	n, err = c.outBuf.Read(b)
	if err == io.EOF {
		return 0, nil
	}
	return n, err
}

func (c *bufferConn) Close() error {
	select {
	case <-c.closed:
		return io.ErrClosedPipe
	default:
		close(c.closed)
		c.inBuf.Reset()
		c.outBuf.Reset()
		return nil
	}
}

func (c *bufferConn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *bufferConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *bufferConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *bufferConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *bufferConn) SetWriteDeadline(t time.Time) error {
	return nil
}
