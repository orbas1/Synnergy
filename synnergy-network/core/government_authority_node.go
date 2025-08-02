package core

import (
	"context"

	Nodes "synnergy-network/core/Nodes"
)

// GovernmentAuthorityNodeInterface defines operations for specialised nodes
// that maintain regulatory compliance and legal enforcement on chain.
type GovernmentAuthorityNodeInterface interface {
	Nodes.NodeInterface
	CheckCompliance(tx *Transaction) error
	EnforceRegulation(txID Hash) error
	InterfaceRegulator(data []byte) error
	UpdateLegalFramework(blob []byte) error
	AuditTrail(addr Address) ([]byte, error)
}

// GovernmentAuthorityNode provides integration between the networking layer,
// ledger and compliance engine to fulfil government oversight requirements.
type GovernmentAuthorityNode struct {
	*BaseNode
	ledger *Ledger
	comp   *ComplianceEngine
	ctx    context.Context
	cancel context.CancelFunc
}

// NewGovernmentAuthorityNode wires the required components together.
func NewGovernmentAuthorityNode(net Nodes.NodeInterface, led *Ledger, comp *ComplianceEngine) *GovernmentAuthorityNode {
	ctx, cancel := context.WithCancel(context.Background())
	base := NewBaseNode(net)
	return &GovernmentAuthorityNode{BaseNode: base, ledger: led, comp: comp, ctx: ctx, cancel: cancel}
}

// Close gracefully shuts down the node.
func (g *GovernmentAuthorityNode) Close() error { g.cancel(); return g.BaseNode.Close() }

// CheckCompliance runs a compliance check on a transaction and halts it if risk detected.
func (g *GovernmentAuthorityNode) CheckCompliance(tx *Transaction) error {
	if g.comp == nil {
		return nil
	}
	score, err := g.comp.MonitorTransaction(tx, 0.7)
	if err == nil && score > 0 {
		g.EnforceRegulation(tx.Hash)
	}
	return err
}

// EnforceRegulation marks a transaction as frozen pending review.
func (g *GovernmentAuthorityNode) EnforceRegulation(txID Hash) error {
	if g.ledger == nil {
		return nil
	}
	return g.ledger.FlagTransaction(txID[:], []byte("frozen"))
}

// InterfaceRegulator is a placeholder for off-chain regulator interaction.
func (g *GovernmentAuthorityNode) InterfaceRegulator(data []byte) error {
	// In real deployment this would push data to a government service.
	return nil
}

// UpdateLegalFramework stores new regulatory rules on ledger.
func (g *GovernmentAuthorityNode) UpdateLegalFramework(blob []byte) error {
	if g.ledger == nil {
		return nil
	}
	g.ledger.SetMetadata("regulations", blob)
	return nil
}

// AuditTrail retrieves stored audit data for an address.
func (g *GovernmentAuthorityNode) AuditTrail(addr Address) ([]byte, error) {
	if g.ledger == nil {
		return nil, nil
	}
	return g.ledger.GetMetadata("audit:" + addr.Hex())
}

// Ensure GovernmentAuthorityNode satisfies the interface
var _ GovernmentAuthorityNodeInterface = (*GovernmentAuthorityNode)(nil)
