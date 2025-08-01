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

// HybridNodeInterface extends NodeInterface with multi-role helpers.
type HybridNodeInterface interface {
	NodeInterface
	IndexData(key string, data []byte)
	QueryIndex(key string) ([]byte, bool)
	ProcessTransaction(tx []byte) error
	ProposeBlock() error
}
