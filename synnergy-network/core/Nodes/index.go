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

// AutonomousAgent defines additional behaviour for autonomous nodes.
type AutonomousAgent interface {
	NodeInterface
	AddRule(rule interface{})
	RemoveRule(id string)
	Start()
	Stop() error
// HolographicNodeInterface extends NodeInterface with holographic functions.
type HolographicNodeInterface interface {
	NodeInterface
	EncodeStore(data []byte) (interface{}, error)
	Retrieve(id interface{}) ([]byte, error)
	SyncConsensus(c Consensus) error
	ProcessTx(tx interface{}) error
	ExecuteContract(ctx interface{}, vm VMExecutor, code []byte) error
}

// Ensure the implementation satisfies the interface.
var _ HolographicNodeInterface = (*HolographicNode)(nil)
// TimeLockRecord mirrors core.TimeLockRecord without importing the core package.
type TimeLockRecord struct {
	ID        string
	TokenID   uint32
	From      [20]byte
	To        [20]byte
	Amount    uint64
	ExecuteAt int64
}

// TimeLockedNodeInterface exposes time locked execution features.
type TimeLockedNodeInterface interface {
	NodeInterface
	Queue(TimeLockRecord) error
	Cancel(id string) error
	ExecuteDue() []string
	List() []TimeLockRecord
// EnvironmentalMonitoringInterface extends NodeInterface with sensor management
// and conditional triggers.
type EnvironmentalMonitoringInterface interface {
	NodeInterface
	RegisterSensor(id, endpoint string) error
	RemoveSensor(id string) error
	ListSensors() ([]string, error)
	AddTrigger(id string, threshold float64, action string) error
	Start()
	Stop() error
// MolecularNodeFactory returns a MolecularNodeInterface. Actual constructor lives
// in the core package.
type MolecularNodeFactory func(cfg interface{}) (MolecularNodeInterface, error)
// BiometricSecurityNode extends NodeInterface with biometric operations.
type BiometricSecurityNode interface {
	NodeInterface
	Enroll(addr string, data []byte) error
	Verify(addr string, data []byte) bool
	Delete(addr string)
	ValidateTransaction(tx any, data []byte) bool
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
