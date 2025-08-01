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

// ValidatorNodeInterface extends NodeInterface with validator specific actions.
// It provides hooks for toggling individual consensus mechanisms and for
// participating in block production.
type ValidatorNodeInterface interface {
	NodeInterface
	EnablePoH(bool)
	EnablePoS(bool)
	EnablePoW(bool)
	Start()
	Stop() error
	ValidateTx([]byte) error
	ProposeBlock() error
	VoteBlock([]byte, []byte) error
}
