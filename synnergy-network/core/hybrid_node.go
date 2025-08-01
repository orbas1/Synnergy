package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// HybridConfig aggregates configuration for the HybridNode.
type HybridConfig struct {
	Network Config
	Ledger  LedgerConfig
}

// HybridNode combines networking, ledger, transaction and indexing services.
type HybridNode struct {
	*NodeAdapter

	ledger    *Ledger
	txPool    *TxPool
	index     map[string][]byte
	consensus *SynnergyConsensus

	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// NewHybridNode creates a hybrid node with its own ledger and tx pool.
func NewHybridNode(cfg *HybridConfig) (*HybridNode, error) {
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
	hn := &HybridNode{
		NodeAdapter: &NodeAdapter{n},
		ledger:      led,
		txPool:      NewTxPool(nil, led, nil, NewFlatGasCalculator(), n, 0),
		index:       make(map[string][]byte),
		ctx:         ctx,
		cancel:      cancel,
	}
	return hn, nil
}

// Start launches the underlying network services.
func (h *HybridNode) Start() { h.NodeAdapter.ListenAndServe() }

// Stop terminates services and closes resources.
func (h *HybridNode) Stop() error {
	h.cancel()
	return h.NodeAdapter.Close()
}

// IndexData stores arbitrary data in the local index.
func (h *HybridNode) IndexData(key string, data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.index == nil {
		h.index = make(map[string][]byte)
	}
	h.index[key] = data
}

// QueryIndex retrieves previously indexed data.
func (h *HybridNode) QueryIndex(key string) ([]byte, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	d, ok := h.index[key]
	return d, ok
}

// ProcessTransaction decodes and adds a transaction to the pool.
func (h *HybridNode) ProcessTransaction(blob []byte) error {
	var tx Transaction
	if err := json.Unmarshal(blob, &tx); err != nil {
		return err
	}
	return h.txPool.AddTx(&tx)
}

// ProposeBlock delegates to the consensus engine if configured.
func (h *HybridNode) ProposeBlock() (*SubBlock, error) {
	if h.consensus == nil {
		return nil, fmt.Errorf("consensus not configured")
	}
	return h.consensus.ProposeSubBlock()
}

// Ledger exposes the node ledger.
func (h *HybridNode) Ledger() *Ledger { return h.ledger }
