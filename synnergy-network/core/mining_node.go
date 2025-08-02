package core

import (
	"context"

	"github.com/sirupsen/logrus"
)

// MiningNode bundles networking, ledger and consensus components for PoW mining.
type MiningNode struct {
	*BaseNode
	ledger *Ledger
	cons   *SynnergyConsensus
	pool   *TxPool
	ctx    context.Context
	cancel context.CancelFunc
}

// MiningNodeConfig aggregates the required configurations.
type MiningNodeConfig struct {
	Network Config
	Ledger  LedgerConfig
}

// NewMiningNode constructs a mining node with networking, ledger and consensus
// services wired together. It returns the ready-to-start node.
func NewMiningNode(cfg *MiningNodeConfig) (*MiningNode, error) {
	ctx, cancel := context.WithCancel(context.Background())

	n, err := NewNode(cfg.Network)
	if err != nil {
		cancel()
		return nil, err
	}

	led, err := NewLedger(cfg.Ledger)
	if err != nil {
		cancel()
		_ = n.Close()
		return nil, err
	}

	pool := NewTxPool(nil, led, nil, nil, 0)

	cons, err := NewConsensus(logrus.New(), led, n, nil, pool, nil)
	if err != nil {
		cancel()
		_ = n.Close()
		return nil, err
	}

	base := NewBaseNode(&NodeAdapter{n})
	return &MiningNode{BaseNode: base, ledger: led, cons: cons, pool: pool, ctx: ctx, cancel: cancel}, nil
}

// StartMining launches the networking and consensus loops.
func (m *MiningNode) StartMining() {
	go m.ListenAndServe()
	m.cons.Start(m.ctx)
}

// StopMining gracefully shuts down the mining node.
func (m *MiningNode) StopMining() error {
	m.cancel()
	return m.Close()
}

// AddTransaction validates and queues a transaction for inclusion in a block.
func (m *MiningNode) AddTransaction(tx *Transaction) error {
	return m.pool.AddTx(tx)
}
