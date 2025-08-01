package core

import (
	"context"
	"encoding/json"
	Nodes "synnergy-network/core/Nodes"
)

// RegulatoryConfig aggregates config required to bootstrap a regulatory node.
type RegulatoryConfig struct {
	Network        Config
	Ledger         LedgerConfig
	TrustedIssuers [][]byte
}

// RegulatoryNode enforces compliance rules and exposes network services.
type RegulatoryNode struct {
	net       *Node
	ledger    *Ledger
	consensus *SynnergyConsensus
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewRegulatoryNode initialises networking, ledger and compliance subsystems.
func NewRegulatoryNode(cfg *RegulatoryConfig) (*RegulatoryNode, error) {
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
	InitRegulatory(led)
	InitCompliance(led, cfg.TrustedIssuers)
	rn := &RegulatoryNode{net: n, ledger: led, ctx: ctx, cancel: cancel}
	return rn, nil
}

// Start launches the underlying network services.
func (r *RegulatoryNode) Start() { go r.net.ListenAndServe() }

// Stop gracefully shuts down the node.
func (r *RegulatoryNode) Stop() error {
	r.cancel()
	return r.net.Close()
}

// DialSeed proxies to the underlying network node.
func (r *RegulatoryNode) DialSeed(peers []string) error { return r.net.DialSeed(peers) }

// Broadcast sends data to peers using the underlying network node.
func (r *RegulatoryNode) Broadcast(topic string, data []byte) error {
	return r.net.Broadcast(topic, data)
}

// Subscribe wraps the underlying node subscription and exposes raw byte messages.
func (r *RegulatoryNode) Subscribe(topic string) (<-chan []byte, error) {
	ch, err := r.net.Subscribe(topic)
	if err != nil {
		return nil, err
	}
	out := make(chan []byte)
	go func() {
		for msg := range ch {
			out <- msg.Data
		}
	}()
	return out, nil
}

// ListenAndServe starts listening for network traffic.
func (r *RegulatoryNode) ListenAndServe() { r.net.ListenAndServe() }

// Close stops the node.
func (r *RegulatoryNode) Close() error { return r.Stop() }

// Peers returns the list of connected peer IDs.
func (r *RegulatoryNode) Peers() []string { return r.net.Peers() }

// Ledger exposes the underlying ledger.
func (r *RegulatoryNode) Ledger() *Ledger { return r.ledger }

// SetConsensus attaches a consensus engine for advanced integrations.
func (r *RegulatoryNode) SetConsensus(c *SynnergyConsensus) { r.consensus = c }

// VerifyTransaction enforces regulatory rule-set on the given transaction.
func (r *RegulatoryNode) VerifyTransaction(tx *Transaction) error {
	return EvaluateRuleSet(tx)
}

// VerifyKYC records a validated KYC document via the compliance engine.
func (r *RegulatoryNode) VerifyKYC(doc *KYCDocument) error {
	return Compliance().ValidateKYC(doc)
}

// EraseKYC removes stored personal data while keeping commitments.
func (r *RegulatoryNode) EraseKYC(addr Address) error { return Compliance().EraseData(addr) }

// RiskScore returns the fraud score for an address.
func (r *RegulatoryNode) RiskScore(addr Address) int { return Compliance().RiskScore(addr) }

// GenerateReport exports the list of registered regulators as JSON.
func (r *RegulatoryNode) GenerateReport() ([]byte, error) {
	regs := ListRegulators()
	return json.Marshal(regs)
}

// Compile-time assertions
var _ Nodes.NodeInterface = (*RegulatoryNode)(nil)
