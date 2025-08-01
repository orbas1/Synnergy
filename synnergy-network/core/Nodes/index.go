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

// EnergyNodeInterface extends NodeInterface with energy tracking methods.
type EnergyNodeInterface interface {
	NodeInterface
	RecordUsage(txs uint64, kwh float64) error
	Efficiency() (float64, error)
	NetworkAverage() (float64, error)
}
