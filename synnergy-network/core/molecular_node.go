package core

import (
	"fmt"
	"sync"

	Nodes "synnergy-network/core/Nodes"
)

// MolecularNode operates at the molecular level combining networking with ledger
// integration. This is a speculative prototype for future nano-scale nodes.
type MolecularNode struct {
	net    *Node
	ledger *Ledger
	mu     sync.RWMutex
}

// MolecularNodeConfig aggregates network and ledger settings.
type MolecularNodeConfig struct {
	Network Config
	Ledger  LedgerConfig
}

// NewMolecularNode constructs a MolecularNode with networking and ledger
// components. It returns a node ready to start.
func NewMolecularNode(cfg MolecularNodeConfig) (*MolecularNode, error) {
	n, err := NewNode(cfg.Network)
	if err != nil {
		return nil, err
	}
	led, err := NewLedger(cfg.Ledger)
	if err != nil {
		_ = n.Close()
		return nil, err
	}
	return &MolecularNode{net: n, ledger: led}, nil
}

// DialSeed proxies to the underlying network node.
func (m *MolecularNode) DialSeed(peers []string) error { return m.net.DialSeed(peers) }

// Broadcast proxies network broadcast.
func (m *MolecularNode) Broadcast(topic string, data []byte) error {
	return m.net.Broadcast(topic, data)
}

// Subscribe proxies network subscription and converts messages.
func (m *MolecularNode) Subscribe(topic string) (<-chan []byte, error) {
	ch, err := m.net.Subscribe(topic)
	if err != nil {
		return nil, err
	}
	out := make(chan []byte)
	go func() {
		for msg := range ch {
			out <- msg.Data
		}
	}()
	return out, nil
}

// ListenAndServe starts the underlying network node.
func (m *MolecularNode) ListenAndServe() { m.net.ListenAndServe() }

// Close shuts down network and ledger.
func (m *MolecularNode) Close() error {
	err := m.net.Close()
	return err
}

// Peers lists known peers as strings.
func (m *MolecularNode) Peers() []string {
	peers := m.net.Peers()
	list := make([]string, len(peers))
	for i, p := range peers {
		list[i] = fmt.Sprintf("%s", p.ID)
	}
	return list
}

// AtomicTransaction submits a transaction to the ledger pool.
func (m *MolecularNode) AtomicTransaction(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ledger == nil {
		return fmt.Errorf("ledger not initialised")
	}
	tx := &Transaction{Payload: data}
	m.ledger.AddToPool(tx)
	return nil
}

// EncodeDataInMatter is a stub demonstrating future data storage in matter.
func (m *MolecularNode) EncodeDataInMatter(data []byte) (string, error) {
	return fmt.Sprintf("molecular:%x", data), nil
}

// MonitorNanoSensors returns dummy sensor data.
func (m *MolecularNode) MonitorNanoSensors() ([]byte, error) { return []byte("ok"), nil }

// ControlMolecularProcess accepts a command payload. Currently a stub.
func (m *MolecularNode) ControlMolecularProcess(cmd []byte) error { return nil }

// Ensure MolecularNode implements the interface.
var _ Nodes.MolecularNodeInterface = (*MolecularNode)(nil)
