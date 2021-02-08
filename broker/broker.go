package broker

import (
	"bufio"
	"log"
	"net"
	"sync"
)

// Broker type
type Broker struct {
	connPool map[string]net.Conn
	cpMtx    sync.RWMutex
}

// New create a new Broker
func New() *Broker {
	return &Broker{
		connPool: make(map[string]net.Conn, 100),
	}
}

// Serve serves connections from the listener
func (b *Broker) Serve(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		log.Printf("Connected: %s\n", conn.RemoteAddr())
		b.storeConn(conn.RemoteAddr().String(), conn)
		go b.serveConn(conn)
	}
}

func (b *Broker) serveConn(conn net.Conn) {
	defer b.onDisconnect(conn)
	r := bufio.NewReader(conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			break
		}
		log.Printf("%s >> %s", conn.RemoteAddr(), msg)
		b.broadcast(msg)
	}
}

func (b *Broker) onDisconnect(conn net.Conn) {
	log.Printf("Disconnected: %s\n", conn.RemoteAddr())
	b.deleteConn(conn.RemoteAddr().String())
}

func (b *Broker) broadcast(msg string) {
	conns := b.allConns()
	for _, conn := range conns {
		conn.Write([]byte(msg))
	}
}

func (b *Broker) allConns() []net.Conn {
	b.cpMtx.RLock()
	defer b.cpMtx.RUnlock()
	ret := make([]net.Conn, 0, len(b.connPool))
	for _, conn := range b.connPool {
		ret = append(ret, conn)
	}
	return ret
}

func (b *Broker) loadConn(key string) net.Conn {
	b.cpMtx.RLock()
	defer b.cpMtx.RUnlock()
	return b.connPool[key]
}

func (b *Broker) storeConn(key string, conn net.Conn) {
	b.cpMtx.Lock()
	defer b.cpMtx.Unlock()
	b.connPool[key] = conn
}

func (b *Broker) deleteConn(key string) {
	b.cpMtx.Lock()
	defer b.cpMtx.Unlock()
	delete(b.connPool, key)
}
