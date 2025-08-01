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

// CentralBankingNode defines the extended behaviour required by central bank
// infrastructure. It mirrors the high level actions without importing core
// types to avoid circular dependencies.
type CentralBankingNode interface {
	NodeInterface

	// Monetary policy controls
	SetInterestRate(float64) error
	InterestRate() float64
	SetReserveRequirement(float64) error
	ReserveRequirement() float64

	// Digital currency issuance and settlement hooks. Addresses and
	// transactions are passed as raw bytes to keep this package decoupled.
	IssueDigitalCurrency(addr [20]byte, amount uint64) error
	RecordSettlement(tx []byte) error
}
