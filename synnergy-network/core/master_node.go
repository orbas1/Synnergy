package core

import (
	"errors"
	"sync"
	Nodes "synnergy-network/core/Nodes"
)

// MasterNode provides enhanced transaction processing, privacy services and
// governance utilities. It embeds a standard Node to reuse the networking
// stack and attaches references to the ledger and consensus modules.

type MasterNode struct {
	node       *Node
	ledger     *Ledger
	consensus  *SynnergyConsensus
	gov        *SYN300Token
	address    Address
	collateral uint64
	mu         sync.RWMutex
}

// NewMasterNode constructs a MasterNode from its dependencies. The caller is
// responsible for supplying a configured network Node, ledger and consensus
// instances. The SYN300 governance token is optional but enables on-chain
// voting via VoteProposal.
func NewMasterNode(n *Node, led *Ledger, cons *SynnergyConsensus, gov *SYN300Token, addr Address, collateral uint64) *MasterNode {
	return &MasterNode{
		node:       n,
		ledger:     led,
		consensus:  cons,
		gov:        gov,
		address:    addr,
		collateral: collateral,
	}
}

// -----------------------------------------------------------------------------
// NodeInterface forwarding
// -----------------------------------------------------------------------------

func (m *MasterNode) DialSeed(seeds []string) error                 { return m.node.DialSeed(seeds) }
func (m *MasterNode) Broadcast(topic string, data []byte) error     { return m.node.Broadcast(topic, data) }
func (m *MasterNode) Subscribe(topic string) (<-chan []byte, error) { return m.node.Subscribe(topic) }
func (m *MasterNode) ListenAndServe()                               { m.node.ListenAndServe() }
func (m *MasterNode) Close() error                                  { return m.node.Close() }
func (m *MasterNode) Peers() []string {
	peers := m.node.Peers()
	out := make([]string, len(peers))
	for i, p := range peers {
		out[i] = string(p.ID)
	}
	return out
}

// Start begins the underlying services.
func (m *MasterNode) Start() { go m.ListenAndServe() }

// Stop closes all services.
func (m *MasterNode) Stop() error { return m.Close() }

// -----------------------------------------------------------------------------
// Master node specific functionality
// -----------------------------------------------------------------------------

// ProcessTx inserts the transaction into the ledger pool for expedited
// processing.
func (m *MasterNode) ProcessTx(tx *Transaction) error {
	if tx == nil {
		return errors.New("nil transaction")
	}
	if m.ledger == nil {
		return errors.New("ledger not configured")
	}
	m.ledger.AddToPool(tx)
	return nil
}

// HandlePrivateTx encrypts the payload and submits it to the ledger.
func (m *MasterNode) HandlePrivateTx(tx *Transaction, key []byte) error {
	if err := EncryptTxPayload(tx, key); err != nil {
		return err
	}
	return m.ProcessTx(tx)
}

// VoteProposal casts a governance vote using the SYN300 token if available.
func (m *MasterNode) VoteProposal(id uint64, approve bool) error {
	if m.gov == nil {
		return errors.New("governance token not set")
	}
	m.gov.Vote(id, m.address, approve)
	return nil
}

// Ensure MasterNode implements the interface.
var _ Nodes.MasterNodeInterface = (*MasterNode)(nil)
