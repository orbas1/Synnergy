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

// BiometricSecurityNode extends NodeInterface with biometric operations.
type BiometricSecurityNode interface {
	NodeInterface
	Enroll(addr string, data []byte) error
	Verify(addr string, data []byte) bool
	Delete(addr string)
	ValidateTransaction(tx any, data []byte) bool
}
