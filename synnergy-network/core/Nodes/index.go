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

// RegulatoryNodeInterface extends NodeInterface with compliance helpers.
type RegulatoryNodeInterface interface {
	NodeInterface
	VerifyTransaction([]byte) error
	VerifyKYC([]byte) error
	EraseKYC(string) error
	RiskScore(string) int
	GenerateReport() ([]byte, error)
}
