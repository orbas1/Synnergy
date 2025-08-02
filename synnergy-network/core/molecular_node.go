package core

import (
	"fmt"
	"sync"
)

// MolecularNode operates at the molecular level combining networking with ledger
// integration. This is a speculative prototype for future nano-scale nodes.
type MolecularNode struct {
	*BaseNode
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
	base := NewBaseNode(&NodeAdapter{n})
	return &MolecularNode{BaseNode: base, ledger: led}, nil
}

// Peers lists known peers as strings.
func (m *MolecularNode) Peers() []string {
	peers := m.BaseNode.Peers()
	list := make([]string, len(peers))
	copy(list, peers)
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
