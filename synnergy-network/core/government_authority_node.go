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
	net    Nodes.NodeInterface
	ledger *Ledger
	comp   *ComplianceEngine
	ctx    context.Context
	cancel context.CancelFunc
}

// NewGovernmentAuthorityNode wires the required components together.
func NewGovernmentAuthorityNode(net Nodes.NodeInterface, led *Ledger, comp *ComplianceEngine) *GovernmentAuthorityNode {
	ctx, cancel := context.WithCancel(context.Background())
	return &GovernmentAuthorityNode{net: net, ledger: led, comp: comp, ctx: ctx, cancel: cancel}
}

// ListenAndServe begins processing of network events.
func (g *GovernmentAuthorityNode) ListenAndServe() { g.net.ListenAndServe() }

// Close gracefully shuts down the node.
func (g *GovernmentAuthorityNode) Close() error { g.cancel(); return g.net.Close() }

// DialSeed proxies to the underlying network interface.
func (g *GovernmentAuthorityNode) DialSeed(seeds []string) error { return g.net.DialSeed(seeds) }

// Broadcast proxies to network.
func (g *GovernmentAuthorityNode) Broadcast(topic string, data []byte) error {
	return g.net.Broadcast(topic, data)
}

// Subscribe proxies to network.
func (g *GovernmentAuthorityNode) Subscribe(topic string) (<-chan []byte, error) {
	return g.net.Subscribe(topic)
}

// Peers returns the peer list.
func (g *GovernmentAuthorityNode) Peers() []string { return g.net.Peers() }

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
