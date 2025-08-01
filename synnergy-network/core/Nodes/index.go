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

// BankInstitutionalNode defines behaviour for specialised
// bank/institution authority nodes.
type BankInstitutionalNode interface {
	NodeInterface
	MonitorTransaction(data []byte) error
	ComplianceReport() ([]byte, error)
	ConnectFinancialNetwork(endpoint string) error
	UpdateRuleset(rules map[string]interface{})
// WarfareNodeInterface is implemented by nodes specialised for military
// operations. It embeds NodeInterface and exposes additional methods defined
// in the military_nodes subpackage.
//
// Keeping the interface here avoids package import cycles while allowing the
// core package to rely on the abstract type.
type WarfareNodeInterface interface {
	NodeInterface
	SecureCommand(data []byte) error
	TrackLogistics(itemID, status string) error
	ShareTactical(data []byte) error
// MobileMiner extends NodeInterface with light mining controls.
type MobileMiner interface {
	NodeInterface
	StartMining()
	StopMining()
	SetIntensity(int)
	Stats() any
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
