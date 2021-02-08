package icmpnet

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type secureConn struct {
	*bufferConn
	baseConn net.Conn
	aesgcm   cipher.AEAD
	aesKey   []byte
}

func newSecureConn(baseConn net.Conn, aesKey []byte) (*secureConn, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	sc := &secureConn{
		bufferConn: newBufferConn(baseConn.LocalAddr(), baseConn.RemoteAddr()),
		baseConn:   baseConn,
		aesgcm:     aesgcm,
		aesKey:     aesKey,
	}
	go sc.readLoop()
	go sc.writeLoop()
	return sc, nil
}

func (sc *secureConn) readLoop() {
	defer func() {
		sc.Close()
		sc.baseConn.Close()
	}()

	for {
		sizeB := make([]byte, 4)
		if _, err := io.ReadFull(sc.baseConn, sizeB); err != nil {
			return
		}

		size := binary.BigEndian.Uint32(sizeB)
		if size > 32768 {
			return
		}
		emsg := make([]byte, size)
		if err := sc.readFullTimeout(emsg, 2*time.Second); err != nil {
			return
		}
		msg, err := sc.aesgcm.Open(nil, sc.aesKey[:12], emsg, nil)
		if err != nil {
			return
		}

		sc.writeInBuf(msg)
	}
}

func (sc *secureConn) readFullTimeout(b []byte, d time.Duration) error {
	ch := make(chan error)

	go func() {
		_, err := io.ReadFull(sc.baseConn, b)
		ch <- err
	}()

	select {
	case err := <-ch:
		return err
	case <-time.After(d):
		return fmt.Errorf("read timeout")
	}
}

func (sc *secureConn) writeLoop() {
	defer func() {
		sc.Close()
		sc.baseConn.Close()
	}()

	buf := make([]byte, 32768)
	for {
		n, _ := sc.readOutBuf(buf)
		if n == 0 {
			time.Sleep(time.Millisecond)
			continue
		}
		msg := buf[:n]

		emsg := sc.aesgcm.Seal(nil, sc.aesKey[:12], msg, nil)
		size := uint32(len(emsg))

		sizeB := make([]byte, 4)
		binary.BigEndian.PutUint32(sizeB, size)

		_, err := sc.baseConn.Write(sizeB)
		if err != nil {
			break
		}
		_, err = sc.baseConn.Write(emsg)
		if err != nil {
			break
		}
	}
}
