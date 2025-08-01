package core

import (
	"sync"

	Nodes "synnergy-network/core/Nodes"
)

// IndexingNode provides fast query capabilities by indexing ledger data.
type IndexingNode struct {
	net    *Node
	ledger *Ledger

	mu      sync.RWMutex
	txIndex map[Address][]*Transaction
	// stateIndex stores arbitrary key/value pairs per contract address.
	stateIndex map[Address]map[string][]byte
}

// IndexingNodeConfig contains the network configuration for a new indexing node.
type IndexingNodeConfig struct {
	Network Config
}

// NewIndexingNode constructs a new node with optional initial indexing of the
// provided ledger.
func NewIndexingNode(cfg IndexingNodeConfig, led *Ledger) (*IndexingNode, error) {
	net, err := NewNode(cfg.Network)
	if err != nil {
		return nil, err
	}
	idx := &IndexingNode{
		net:        net,
		ledger:     led,
		txIndex:    make(map[Address][]*Transaction),
		stateIndex: make(map[Address]map[string][]byte),
	}
	if led != nil {
		for _, blk := range led.Blocks {
			idx.indexBlock(blk)
		}
	}
	return idx, nil
}

// Ledger returns the bound ledger instance.
func (i *IndexingNode) Ledger() *Ledger { return i.ledger }

// indexBlock updates the internal indexes based on block contents.
func (i *IndexingNode) indexBlock(b *Block) {
	for _, tx := range b.Transactions {
		i.txIndex[tx.From] = append(i.txIndex[tx.From], tx)
		i.txIndex[tx.To] = append(i.txIndex[tx.To], tx)
	}
}

// AddBlock indexes a new block. Call this after the block has been applied to
// the ledger.
func (i *IndexingNode) AddBlock(b *Block) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.indexBlock(b)
}

// QueryTxHistory returns a copy of all transactions involving the address.
func (i *IndexingNode) QueryTxHistory(addr Address) []*Transaction {
	i.mu.RLock()
	defer i.mu.RUnlock()
	txs := i.txIndex[addr]
	out := make([]*Transaction, len(txs))
	copy(out, txs)
	return out
}

// RecordState stores key/value data for a contract address.
func (i *IndexingNode) RecordState(addr Address, key string, val []byte) {
	i.mu.Lock()
	defer i.mu.Unlock()
	m := i.stateIndex[addr]
	if m == nil {
		m = make(map[string][]byte)
		i.stateIndex[addr] = m
	}
	m[key] = val
}

// QueryState retrieves previously recorded state data.
func (i *IndexingNode) QueryState(addr Address, key string) ([]byte, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	m := i.stateIndex[addr]
	if m == nil {
		return nil, false
	}
	val, ok := m[key]
	return val, ok
}

// ---- Network integration ----

func (i *IndexingNode) DialSeed(peers []string) error { return i.net.DialSeed(peers) }
func (i *IndexingNode) Broadcast(topic string, data []byte) error {
	return i.net.Broadcast(topic, data)
}
func (i *IndexingNode) Subscribe(topic string) (<-chan []byte, error) { return i.net.Subscribe(topic) }
func (i *IndexingNode) ListenAndServe()                               { i.net.ListenAndServe() }
func (i *IndexingNode) Close() error                                  { return i.net.Close() }
func (i *IndexingNode) Peers() []string                               { return i.net.Peers() }

// Ensure IndexingNode implements the shared interface.
var _ Nodes.NodeInterface = (*IndexingNode)(nil)
