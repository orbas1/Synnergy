package core

import Nodes "synnergy-network/core/Nodes"

// IntegrationNode extends a network node with facilities to track external APIs
// and blockchain bridges. The struct lives in the core package to avoid import
// cycles with the Nodes subpackage.
// IntegrationNode represents a network participant that can interface with
// external APIs and other blockchains. JSON tags are provided for clarity when
// serialising this structure.
type IntegrationNode struct {
	Nodes.NodeInterface `json:"-"`
	Ledger              *Ledger              `json:"ledger"`
	Registry            *IntegrationRegistry `json:"registry"`
}

// NewIntegrationNode creates a new instance using the provided network node and
// ledger. If reg is nil a fresh registry is used.
func NewIntegrationNode(n Nodes.NodeInterface, led *Ledger, reg *IntegrationRegistry) *IntegrationNode {
	if reg == nil {
		reg = NewIntegrationRegistry()
	}
	return &IntegrationNode{NodeInterface: n, Ledger: led, Registry: reg}
}

// RelayTransaction places a transaction onto the local ledger pool.
func (in *IntegrationNode) RelayTransaction(tx *Transaction) error {
	if in.Ledger == nil || tx == nil {
		return nil
	}
	in.Ledger.AddToPool(tx)
	return nil
}

// RegisterAPI adds an external API endpoint to the registry.
func (in *IntegrationNode) RegisterAPI(name, endpoint string) error {
	in.Registry.RegisterAPI(name, endpoint)
	return nil
}

// RemoveAPI removes an API from the registry.
func (in *IntegrationNode) RemoveAPI(name string) error {
	in.Registry.RemoveAPI(name)
	return nil
}

// ListAPIs returns names of registered APIs.
func (in *IntegrationNode) ListAPIs() []string { return in.Registry.ListAPIs() }

// ConnectChain registers a connection to another blockchain.
func (in *IntegrationNode) ConnectChain(id, endpoint string) error {
	in.Registry.ConnectChain(id, endpoint)
	return nil
}

// DisconnectChain drops a chain connection.
func (in *IntegrationNode) DisconnectChain(id string) error {
	in.Registry.DisconnectChain(id)
	return nil
}

// ListChains lists connected chains.
func (in *IntegrationNode) ListChains() []string { return in.Registry.ListChains() }
