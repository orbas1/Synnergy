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

// MobileMiner extends NodeInterface with light mining controls.
type MobileMiner interface {
	NodeInterface
	StartMining()
	StopMining()
	SetIntensity(int)
	Stats() any
}
