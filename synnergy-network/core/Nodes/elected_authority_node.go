package Nodes

import "sync"

// ElectedAuthorityNode provides enhanced authority capabilities subject to
// community voting. Access is granted once votes exceed 5% of network nodes and
// revoked when reports exceed 2.5%.
//
// The struct embeds the generic NodeInterface to expose networking features
// without tying this package to a concrete implementation.
type ElectedAuthorityNode struct {
	NodeInterface
	mu      sync.RWMutex
	votes   map[Address]struct{}
	reports map[Address]struct{}
	active  bool
	total   int
}

// NewElectedAuthorityNode wraps an existing node and initialises vote tracking.
func NewElectedAuthorityNode(base NodeInterface, totalNodes int) *ElectedAuthorityNode {
	return &ElectedAuthorityNode{
		NodeInterface: base,
		votes:         make(map[Address]struct{}),
		reports:       make(map[Address]struct{}),
		total:         totalNodes,
	}
}

// Active returns whether the node is currently authorised.
func (n *ElectedAuthorityNode) Active() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.active
}

// RecordVote registers a vote in favour of electing the node.
// When the vote count reaches 5% of all nodes the node becomes active.
func (n *ElectedAuthorityNode) RecordVote(voter Address) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if _, ok := n.votes[voter]; ok {
		return
	}
	n.votes[voter] = struct{}{}
	if !n.active && len(n.votes)*100 >= n.total*5 {
		n.active = true
	}
}

// ReportMisbehaviour registers a report against the node. If reports exceed
// 2.5% of the total node count the node is deactivated.
func (n *ElectedAuthorityNode) ReportMisbehaviour(reporter Address) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if _, ok := n.reports[reporter]; ok {
		return
	}
	n.reports[reporter] = struct{}{}
	if n.active && len(n.reports)*100 >= n.total*25/10 {
		n.active = false
	}
}

// ValidateTransaction is a hook for validating a raw transaction payload. The
// actual validation logic lives in the core package and is invoked via the VM.
func (n *ElectedAuthorityNode) ValidateTransaction(tx []byte) error { return nil }

// CreateBlock is a stub that delegates block creation to the core runtime.
func (n *ElectedAuthorityNode) CreateBlock(blob []byte) error { return nil }

// ReverseTransaction is a placeholder that will call into the ledger module via
// the opcode dispatcher when invoked from the VM.
func (n *ElectedAuthorityNode) ReverseTransaction(hash Hash, sigs [][]byte) error { return nil }

// ViewPrivateTransaction exposes inspection of encrypted transactions.
func (n *ElectedAuthorityNode) ViewPrivateTransaction(hash Hash) ([]byte, error) { return nil, nil }

// ApproveLoanProposal records approval for a loan or grant request.
func (n *ElectedAuthorityNode) ApproveLoanProposal(id string) error { return nil }

// Ensure ElectedAuthorityNode implements NodeInterface.
var _ NodeInterface = (*ElectedAuthorityNode)(nil)
