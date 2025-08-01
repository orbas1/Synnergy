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

// LightNodeInterface extends NodeInterface with header specific accessors.
// Light nodes only maintain block headers and request full blocks on demand.
type LightNodeInterface interface {
	NodeInterface
	StoreHeader(h BlockHeader)
	Headers() []BlockHeader
}
