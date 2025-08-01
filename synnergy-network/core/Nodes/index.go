package Nodes

// NodeInterface defines minimal node behaviour independent from core types.
type NodeInterface interface {
	DialSeed([]string) error
	Broadcast(topic string, data []byte) error
	Subscribe(topic string) (<-chan []byte, error)
	ListenAndServe()
	Close() error
	Peers() []string
}

// MasterNodeInterface extends NodeInterface with specialised services used by
// Synthron master nodes. The concrete implementation lives in the core package
// to avoid an import cycle.
type MasterNodeInterface interface {
	NodeInterface

	// ProcessTx submits a standard transaction for expedited processing.
	ProcessTx(tx any) error

	// HandlePrivateTx encrypts and submits a privacy preserving transaction.
	HandlePrivateTx(tx any, key []byte) error

	// VoteProposal allows the master node to participate in on-chain
	// governance via the SYN300 token module.
	VoteProposal(id uint64, approve bool) error

	// Start activates the underlying services (network, consensus, etc.).
	Start()

	// Stop gracefully shuts down all services.
	Stop() error
}
