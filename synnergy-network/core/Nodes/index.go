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

// HolographicNodeInterface extends NodeInterface with holographic functions.
type HolographicNodeInterface interface {
	NodeInterface
	EncodeStore(data []byte) (interface{}, error)
	Retrieve(id interface{}) ([]byte, error)
	SyncConsensus(c Consensus) error
	ProcessTx(tx interface{}) error
	ExecuteContract(ctx interface{}, vm VMExecutor, code []byte) error
}

// Ensure the implementation satisfies the interface.
var _ HolographicNodeInterface = (*HolographicNode)(nil)
