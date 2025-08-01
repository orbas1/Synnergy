package Nodes

import (
	"context"
	"sync"
)

// BlockHeader contains minimal block information used by light nodes.
type BlockHeader struct {
	Hash      [32]byte
	Height    uint64
	Previous  [32]byte
	Timestamp int64
}

// LightNode implements a lightweight client that stores only block headers.
type LightNode struct {
	parent  NodeInterface
	headers []BlockHeader
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewLightNode wraps an existing node interface with light node capabilities.
// The parent is typically a network node adapter that provides connectivity.
func NewLightNode(parent NodeInterface) *LightNode {
	ctx, cancel := context.WithCancel(context.Background())
	return &LightNode{
		parent:  parent,
		headers: make([]BlockHeader, 0),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (l *LightNode) DialSeed(peers []string) error { return l.parent.DialSeed(peers) }
func (l *LightNode) Broadcast(topic string, data []byte) error {
	return l.parent.Broadcast(topic, data)
}
func (l *LightNode) Subscribe(topic string) (<-chan []byte, error) { return l.parent.Subscribe(topic) }
func (l *LightNode) ListenAndServe()                               { l.parent.ListenAndServe() }
func (l *LightNode) Close() error                                  { l.cancel(); return l.parent.Close() }
func (l *LightNode) Peers() []string                               { return l.parent.Peers() }

// StoreHeader records a new block header in memory.
func (l *LightNode) StoreHeader(h BlockHeader) {
	l.mu.Lock()
	l.headers = append(l.headers, h)
	l.mu.Unlock()
}

// Headers returns a snapshot of all stored headers.
func (l *LightNode) Headers() []BlockHeader {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]BlockHeader, len(l.headers))
	copy(out, l.headers)
	return out
}

// Ensure LightNode conforms to the generic interface.
var _ NodeInterface = (*LightNode)(nil)
