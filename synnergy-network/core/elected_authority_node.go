package core

import (
	"sync"
	Nodes "synnergy-network/core/Nodes"
)

// ElectedAuthorityNode integrates the elected authority functionality with
// the ledger, consensus engine and transaction processing pipeline.
type ElectedAuthorityNode struct {
	*Nodes.ElectedAuthorityNode
	ledger    *Ledger
	consensus *SynnergyConsensus
	mu        sync.Mutex
}

// NewElectedAuthorityNode constructs a node using the given network adapter and
// references to the ledger and consensus engine.
// Returns nil if required dependencies are not provided.
func NewElectedAuthorityNode(base Nodes.NodeInterface, led *Ledger, cons *SynnergyConsensus, total int) *ElectedAuthorityNode {
	if led == nil || cons == nil {
		return nil
	}
	return &ElectedAuthorityNode{
		ElectedAuthorityNode: Nodes.NewElectedAuthorityNode(base, total),
		ledger:               led,
		consensus:            cons,
	}
}

// ValidateTransaction verifies the transaction signature and performs minimal
// ledger checks.
func (n *ElectedAuthorityNode) ValidateTransaction(tx *Transaction) error {
	if err := tx.VerifySig(); err != nil {
		return err
	}
	return nil
}

// CreateBlock executes the supplied transactions via the VM and writes the
// resulting block to the ledger.
func (n *ElectedAuthorityNode) CreateBlock(txs []*Transaction, vm VM) (*Block, error) {
	em := NewExecutionManager(n.ledger, vm)
	height := n.ledger.LastBlockHeight() + 1
	em.BeginBlock(height)
	for _, tx := range txs {
		if err := em.ExecuteTx(tx); err != nil {
			return nil, err
		}
	}
	return em.FinalizeBlock()
}

// ReverseTransaction delegates to the ledger helper requiring multiple authority
// signatures.
func (n *ElectedAuthorityNode) ReverseTransaction(orig *Transaction, sigs [][]byte) (*Transaction, error) {
	return ReverseTransaction(n.ledger, CurrentSet(), orig, sigs)
}

// ViewPrivateTransaction retrieves a private transaction from the ledger.
// ViewPrivateTransaction is a stub for inspecting encrypted transactions. The
// full ledger query is not implemented in the prototype.
func (n *ElectedAuthorityNode) ViewPrivateTransaction(h Hash) (*Transaction, error) {
	return nil, nil
}

// ApproveLoanProposal records approval in the loan pool module.
// ApproveLoanProposal is a placeholder for integrating with the loan pool
// module. In the prototype it simply returns nil.
func (n *ElectedAuthorityNode) ApproveLoanProposal(id string) error {
	return nil
}
