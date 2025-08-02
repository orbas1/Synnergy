package core

import (
	"context"
	"net"
	"testing"
	"time"
)

// startTestServer starts a TCP server that accepts connections and returns listener and slice of accepted conns.
func startTestServer(t *testing.T) (net.Listener, *[]net.Conn) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	conns := &[]net.Conn{}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			*conns = append(*conns, c)
		}
	}()
	return ln, conns
}

func closeServer(ln net.Listener, conns *[]net.Conn) {
	ln.Close()
	for _, c := range *conns {
		c.Close()
	}
}

func TestConnPoolAcquireReuse(t *testing.T) {
	ln, conns := startTestServer(t)
	defer closeServer(ln, conns)

	d := NewDialer(50*time.Millisecond, 50*time.Millisecond)
	cp := NewConnPool(d, 2, time.Second)
	defer cp.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	c1, err := cp.Acquire(ctx, ln.Addr().String())
	if err != nil {
		t.Fatalf("acquire1: %v", err)
	}
	cp.Release(c1)
	if got := cp.Stats(); got != 1 {
		t.Fatalf("expected 1 idle, got %d", got)
	}

	c2, err := cp.Acquire(ctx, ln.Addr().String())
	if err != nil {
		t.Fatalf("acquire2: %v", err)
	}
	if c1 != c2 {
		t.Fatalf("expected to reuse connection")
	}
	cp.Release(c2)
	if got := cp.Stats(); got != 1 {
		t.Fatalf("expected 1 idle after reuse, got %d", got)
	}
}

func TestConnPoolReaper(t *testing.T) {
	ln, conns := startTestServer(t)
	defer closeServer(ln, conns)

	d := NewDialer(50*time.Millisecond, 50*time.Millisecond)
	idle := 100 * time.Millisecond
	cp := NewConnPool(d, 2, idle)
	defer cp.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c, err := cp.Acquire(ctx, ln.Addr().String())
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	cp.Release(c)
	if got := cp.Stats(); got != 1 {
		t.Fatalf("expected 1 idle, got %d", got)
	}

	time.Sleep(3 * idle)
	if got := cp.Stats(); got != 0 {
		t.Fatalf("expected reaper to close idle connections, got %d", got)
	}
}
