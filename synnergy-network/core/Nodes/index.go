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

// ZKPNodeInterface extends NodeInterface with zero-knowledge proof functions.
type ZKPNodeInterface interface {
	NodeInterface
	GenerateProof(data []byte) ([]byte, error)
	VerifyProof(data, proof []byte) bool
	StoreProof(txID string, proof []byte)
	Proof(txID string) ([]byte, bool)
	SubmitTransaction(tx any, proof []byte) error
}
