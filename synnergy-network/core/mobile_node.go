package core

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MobileNode is a lightweight node designed for mobile devices. It wraps the
// standard Node and maintains an optional transaction queue for offline usage.
type MobileNode struct {
	net   *Node
	led   *Ledger
	queue []*Transaction
	mu    sync.Mutex
	off   bool
}

// MobileConfig aggregates the required configuration for a mobile node.
type MobileConfig struct {
	Network Config
	Ledger  LedgerConfig
}

// NewMobileNode creates a new MobileNode with the given configuration.
func NewMobileNode(cfg *MobileConfig) (*MobileNode, error) {
	net, err := NewNode(cfg.Network)
	if err != nil {
		return nil, err
	}
	led, err := NewLedger(cfg.Ledger)
	if err != nil {
		_ = net.Close()
		return nil, err
	}
	return &MobileNode{net: net, led: led}, nil
}

// Start begins network operations.
func (m *MobileNode) Start() { go m.net.ListenAndServe() }

// Stop gracefully shuts down the node.
func (m *MobileNode) Stop() error { return m.net.Close() }

// DialSeed proxies to the underlying network node.
func (m *MobileNode) DialSeed(peers []string) error { return m.net.DialSeed(peers) }

// Broadcast sends a message on the network.
func (m *MobileNode) Broadcast(topic string, data []byte) error {
	return m.net.Broadcast(topic, data)
}

// Subscribe returns a channel of raw messages for the topic.
func (m *MobileNode) Subscribe(topic string) (<-chan Message, error) {
	return m.net.Subscribe(topic)
}

// Peers proxies the peer list from the network node.
func (m *MobileNode) Peers() []*Peer { return m.net.Peers() }

// QueueTx stores the transaction until the device is online.
func (m *MobileNode) QueueTx(tx *Transaction) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queue = append(m.queue, tx)
}

// FlushTxs submits queued transactions to the ledger.
func (m *MobileNode) FlushTxs() {
	m.mu.Lock()
	q := m.queue
	m.queue = nil
	m.mu.Unlock()
	for _, tx := range q {
		m.led.AddToPool(tx)
	}
}

// SetOffline toggles offline mode.
func (m *MobileNode) SetOffline(v bool) {
	m.mu.Lock()
	m.off = v
	m.mu.Unlock()
}

// Offline reports whether the node is in offline mode.
func (m *MobileNode) Offline() bool {
	m.mu.Lock()
	off := m.off
	m.mu.Unlock()
	return off
}

// SyncLedger triggers a best-effort ledger sync using the replicator if
// available. Mobile nodes use a short context timeout by default.
func (m *MobileNode) SyncLedger(ctx context.Context) error {
	rep := NewReplicator(&ReplicationConfig{MaxConcurrent: 1, ChunksPerSec: 5},
		logrus.New(), m.led, nil)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return rep.Synchronize(ctx)
}
