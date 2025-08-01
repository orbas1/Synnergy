package optimization_nodes

import coreNodes "synnergy-network/core/Nodes"
import core "synnergy-network/core"

// OptimizationAPI defines the exposed functionality of an optimization node.
type OptimizationAPI interface {
	coreNodes.NodeInterface
	OptimizeTransactions([]*core.Transaction) []*core.Transaction
	BalanceLoad(peers []string)
}
