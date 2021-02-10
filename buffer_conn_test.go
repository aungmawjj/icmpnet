package icmpnet

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_bufferConn_Read_ShouldReturnErrorWhenClosed(t *testing.T) {
	assert := assert.New(t)
	conn := newBufferConn(nil, nil)
	conn.Close()
	b := make([]byte, 1)
	n, err := conn.Read(b)
	assert.Error(err)
	assert.Equal(0, n)
}

func Test_bufferConn_Read_ShouldNotReturnZero(t *testing.T) {
	assert := assert.New(t)
	conn := newBufferConn(nil, nil) // empty buffer

	go func() {
		b := make([]byte, 10)
		n, err := conn.Read(b)
		if assert.NoError(err) {
			assert.NotEqual(0, n)
		}
	}()

	select {
	case <-time.After(time.Millisecond):
	}
}

func Test_bufferConn_Read_ShouldBlockOnlyUntilIncomingData(t *testing.T) {
	assert := assert.New(t)
	conn := newBufferConn(nil, nil) // empty buffer

	b := make([]byte, 10)
	result := make(chan int, 1)
	go func() {
		n, _ := conn.Read(b)
		result <- n
	}()

	data := []byte{'a', 'b'}
	select {
	case <-time.After(time.Millisecond):
		conn.writeInBuf(data)
	}

	select {
	case n := <-result:
		assert.Equal(2, n)
		assert.EqualValues(data, b[:2])
	case <-time.After(time.Millisecond):
		assert.Fail("Read should be done.")
	}
}

func Test_bufferConn_Write_ShouldReturnErrorWhenClosed(t *testing.T) {
	assert := assert.New(t)
	conn := newBufferConn(nil, nil)
	conn.Close()
	b := make([]byte, 1)
	n, err := conn.Write(b)
	assert.Error(err)
	assert.Equal(0, n)
}

func Test_bufferConn_Write_ShouldWriteAllBytes(t *testing.T) {
	assert := assert.New(t)
	conn := newBufferConn(nil, nil)

	data := []byte{'a', 'b', 'c'}
	conn.Write(data)

	b := make([]byte, 1)
	n, err := conn.readOutBuf(b)
	if assert.NoError(err) {
		assert.Equal(1, n)
		assert.EqualValues(data[:1], b)
	}

	b = make([]byte, 10)
	n, err = conn.readOutBuf(b)
	if assert.NoError(err) {
		assert.Equal(2, n)
		assert.EqualValues(data[1:], b[:2])
	}
}

func Test_bufferConn_readOutBuf_ShouldReturnZeroAndNilWhenEmpty(t *testing.T) {
	assert := assert.New(t)
	conn := newBufferConn(nil, nil)

	b := make([]byte, 1)
	n, err := conn.readOutBuf(b)
	assert.NoError(err)
	assert.Equal(0, n)
}
