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

// APINodeInterface extends NodeInterface with HTTP API controls.
type APINodeInterface interface {
	NodeInterface
	APINode_Start(addr string) error
	APINode_Stop() error
}
