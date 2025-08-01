package core

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

// Connection represents a pooled network connection.
type pooledConn struct {
	net.Conn
	addr     string
	lastUsed time.Time
}

// ConnPool manages reusable network connections keyed by address.
type ConnPool struct {
	dialer    *Dialer
	mu        sync.Mutex
	conns     map[string][]*pooledConn
	maxIdle   int
	idleTTL   time.Duration
	closing   chan struct{}
	closeOnce sync.Once
}

// NewConnPool creates a connection pool using the supplied Dialer. maxIdle defines
// how many idle connections per address are kept. idleTTL specifies how long a
// connection may remain idle before being closed.
func NewConnPool(d *Dialer, maxIdle int, idleTTL time.Duration) *ConnPool {
	cp := &ConnPool{
		dialer:  d,
		conns:   make(map[string][]*pooledConn),
		maxIdle: maxIdle,
		idleTTL: idleTTL,
		closing: make(chan struct{}),
	}
	go cp.reaper()
	return cp
}

// Acquire returns a connection for addr from the pool or establishes a new one.
func (cp *ConnPool) Acquire(ctx context.Context, addr string) (net.Conn, error) {
	cp.mu.Lock()
	list := cp.conns[addr]
	n := len(list)
	if n > 0 {
		c := list[n-1]
		cp.conns[addr] = list[:n-1]
		cp.mu.Unlock()
		c.lastUsed = time.Now()
		return c, nil
	}
	cp.mu.Unlock()
	if cp.dialer == nil {
		return nil, errors.New("connpool: dialer not configured")
	}
	conn, err := cp.dialer.Dial(ctx, addr)
	if err != nil {
		return nil, err
	}
	return &pooledConn{Conn: conn, addr: addr, lastUsed: time.Now()}, nil
}

// Release returns the connection to the pool. Connections not created via
// Acquire are simply closed.
func (cp *ConnPool) Release(conn net.Conn) {
	pc, ok := conn.(*pooledConn)
	if !ok {
		_ = conn.Close()
		return
	}
	cp.mu.Lock()
	defer cp.mu.Unlock()
	if cp.maxIdle > 0 && len(cp.conns[pc.addr]) < cp.maxIdle {
		pc.lastUsed = time.Now()
		cp.conns[pc.addr] = append(cp.conns[pc.addr], pc)
		return
	}
	_ = pc.Close()
}

// Close closes all connections and stops background cleanup.
func (cp *ConnPool) Close() {
	cp.closeOnce.Do(func() {
		close(cp.closing)
		cp.mu.Lock()
		defer cp.mu.Unlock()
		for _, list := range cp.conns {
			for _, c := range list {
				_ = c.Close()
			}
		}
		cp.conns = make(map[string][]*pooledConn)
	})
}

// Stats returns the total number of idle connections managed by the pool.
func (cp *ConnPool) Stats() int {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	count := 0
	for _, list := range cp.conns {
		count += len(list)
	}
	return count
}

// reaper closes idle connections after the configured TTL.
func (cp *ConnPool) reaper() {
	ticker := time.NewTicker(cp.idleTTL / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cutoff := time.Now().Add(-cp.idleTTL)
			cp.mu.Lock()
			for addr, list := range cp.conns {
				i := 0
				for _, c := range list {
					if c.lastUsed.Before(cutoff) {
						_ = c.Close()
						continue
					}
					list[i] = c
					i++
				}
				cp.conns[addr] = list[:i]
			}
			cp.mu.Unlock()
		case <-cp.closing:
			return
		}
	}
}
