package core

import (
	"context"
	"github.com/sirupsen/logrus"
)

// MiningNode bundles networking, ledger and consensus components for PoW mining.
type MiningNode struct {
	net    *Node
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

	return &MiningNode{net: n, ledger: led, cons: cons, pool: pool, ctx: ctx, cancel: cancel}, nil
}

// StartMining launches the networking and consensus loops.
func (m *MiningNode) StartMining() {
	go m.net.ListenAndServe()
	m.cons.Start(m.ctx)
}

// StopMining gracefully shuts down the mining node.
func (m *MiningNode) StopMining() error {
	m.cancel()
	return m.net.Close()
}

// AddTransaction validates and queues a transaction for inclusion in a block.
func (m *MiningNode) AddTransaction(tx *Transaction) error {
	return m.pool.AddTx(tx)
}

// DialSeed proxies to the underlying network node.
func (m *MiningNode) DialSeed(seeds []string) error                 { return m.net.DialSeed(seeds) }
func (m *MiningNode) Broadcast(topic string, data []byte) error     { return m.net.Broadcast(topic, data) }
func (m *MiningNode) Subscribe(topic string) (<-chan []byte, error) { return m.net.Subscribe(topic) }
func (m *MiningNode) ListenAndServe()                               { m.net.ListenAndServe() }
func (m *MiningNode) Close() error                                  { return m.net.Close() }
func (m *MiningNode) Peers() []string                               { return m.net.Peers() }
