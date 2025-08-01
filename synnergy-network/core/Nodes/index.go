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

// QuantumNodeInterface extends NodeInterface with quantum-safe operations.
type QuantumNodeInterface interface {
	NodeInterface
	SecureBroadcast(topic string, data []byte) error
	SecureSubscribe(topic string) (<-chan []byte, error)
	Sign(msg []byte) ([]byte, error)
	Verify(msg, sig []byte) (bool, error)
	RotateKeys() error
}
