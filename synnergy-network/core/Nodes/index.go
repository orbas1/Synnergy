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

// MiningNodeInterface extends NodeInterface with mining specific controls.
type MiningNodeInterface interface {
	NodeInterface
	StartMining()
	StopMining() error
	AddTransaction(tx []byte) error
}
