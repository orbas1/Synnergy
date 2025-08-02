package optimization_nodes

import coreNodes "synnergy-network/core/Nodes"

// Transaction represents the minimal transaction information required for
// optimisation without importing the core package. Only the gas price is
// needed to order transactions.
type Transaction struct {
	GasPrice uint64
}

// Ledger defines the small portion of ledger functionality the optimisation
// node relies on. Decoupling via this interface prevents an import cycle with
// the core package.
type Ledger interface {
	ListPool(limit int) []*Transaction
}

// OptimizationAPI defines the exposed functionality of an optimization node.
// It operates on the lightweight Transaction type to avoid pulling in the
// heavyweight core dependencies, eliminating potential circular imports.
type OptimizationAPI interface {
	coreNodes.NodeInterface
	OptimizeTransactions([]*Transaction) []*Transaction
	BalanceLoad(peers []string)
}
