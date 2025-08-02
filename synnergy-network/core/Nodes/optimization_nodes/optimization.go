package optimization_nodes

import (
	"sort"
	"sync"

	coreNodes "synnergy-network/core/Nodes"
)

// OptimizationNode enhances network performance through transaction ordering
// and basic load balancing.
type OptimizationNode struct {
	base   coreNodes.NodeInterface
	ledger Ledger
	mu     sync.Mutex
}

// NewOptimizationNode creates a node using an existing network node and ledger.
func NewOptimizationNode(base coreNodes.NodeInterface, led Ledger) *OptimizationNode {
	return &OptimizationNode{base: base, ledger: led}
}

// DialSeed proxies to the underlying node.
func (o *OptimizationNode) DialSeed(peers []string) error { return o.base.DialSeed(peers) }

// Broadcast proxies to the underlying node.
func (o *OptimizationNode) Broadcast(t string, d []byte) error { return o.base.Broadcast(t, d) }

// Subscribe proxies to the underlying node.
func (o *OptimizationNode) Subscribe(t string) (<-chan []byte, error) { return o.base.Subscribe(t) }

// ListenAndServe proxies to the underlying node.
func (o *OptimizationNode) ListenAndServe() { o.base.ListenAndServe() }

// Close proxies to the underlying node.
func (o *OptimizationNode) Close() error { return o.base.Close() }

// Peers proxies to the underlying node.
func (o *OptimizationNode) Peers() []string { return o.base.Peers() }

// OptimizeTransactions orders transactions by gas price descending.
func (o *OptimizationNode) OptimizeTransactions(txs []*Transaction) []*Transaction {
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].GasPrice > txs[j].GasPrice
	})
	return txs
}

// BalanceLoad is a stub for dynamic load balancing across peers.
func (o *OptimizationNode) BalanceLoad(peers []string) {
	// Placeholder: actual implementation would monitor peer latency and
	// distribute work accordingly.
}

// ProcessPool fetches transactions from the ledger pool and optimizes them.
func (o *OptimizationNode) ProcessPool(limit int) []*Transaction {
	if o.ledger == nil {
		return nil
	}
	txs := o.ledger.ListPool(limit)
	return o.OptimizeTransactions(txs)
}

var _ OptimizationAPI = (*OptimizationNode)(nil)
