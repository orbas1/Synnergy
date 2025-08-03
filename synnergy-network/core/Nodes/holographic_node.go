package Nodes

import "sync"

// HolographicNode provides holographic data distribution and redundancy.
type HolographicNode struct {
	net    NodeInterface
	ledger Ledger
	mu     sync.RWMutex
}

// Forward NodeInterface behaviour to the embedded node.
func (h *HolographicNode) DialSeed(peers []string) error { return h.net.DialSeed(peers) }
func (h *HolographicNode) Broadcast(topic string, data []byte) error {
	return h.net.Broadcast(topic, data)
}
func (h *HolographicNode) Subscribe(topic string) (<-chan []byte, error) {
	return h.net.Subscribe(topic)
}
func (h *HolographicNode) ListenAndServe() { h.net.ListenAndServe() }
func (h *HolographicNode) Close() error    { return h.net.Close() }
func (h *HolographicNode) Peers() []string { return h.net.Peers() }

// Ledger abstracts ledger interactions required by the holographic node.
type Ledger interface {
	StoreHolographicData([]byte) (interface{}, error)
	HolographicData(id interface{}) ([]byte, error)
	ApplyTransaction(tx interface{}) error
}

// Consensus minimal interface for consensus integration.
type Consensus interface {
	RegisterNode(NodeInterface) error
}

// VMExecutor defines the contract execution interface.
type VMExecutor interface {
	Execute(ctx interface{}, code []byte) error
}

// NewHolographicNode wires the network and ledger into a holographic node.
func NewHolographicNode(net NodeInterface, ledger Ledger) *HolographicNode {
	return &HolographicNode{net: net, ledger: ledger}
}

// Start begins serving on the underlying node.
func (h *HolographicNode) HoloStart() {
	go h.net.ListenAndServe()
}

// HoloStop terminates the underlying node.
func (h *HolographicNode) HoloStop() error {
	return h.net.Close()
}

// EncodeStore holographically encodes data and records it on the ledger.
func (h *HolographicNode) EncodeStore(data []byte) (interface{}, error) {
	return h.ledger.StoreHolographicData(data)
}

// Retrieve decodes previously stored holographic data.
func (h *HolographicNode) Retrieve(id interface{}) ([]byte, error) {
	enc, err := h.ledger.HolographicData(id)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

// SyncConsensus hooks the node into the consensus engine for replication.
func (h *HolographicNode) SyncConsensus(c Consensus) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return c.RegisterNode(h.net)
}

// ProcessTx processes a transaction via the ledger.
func (h *HolographicNode) ProcessTx(tx interface{}) error {
	return h.ledger.ApplyTransaction(tx)
}

// ExecuteContract executes a smart contract on the VM using stored data.
func (h *HolographicNode) ExecuteContract(ctx interface{}, vm VMExecutor, code []byte) error {
	return vm.Execute(ctx, code)
}

// Ensure HolographicNode implements NodeInterface.
var _ NodeInterface = (*HolographicNode)(nil)
