package Nodes

import "context"
 "time"

// NodeInterface defines minimal node behaviour independent from core types.
type NodeInterface interface {
	DialSeed([]string) error
	Broadcast(topic string, data []byte) error
	Subscribe(topic string) (<-chan []byte, error)
	ListenAndServe()
	Close() error
	Peers() []string
}

// ForensicNodeInterface extends NodeInterface with forensic analysis helpers.
// Implementations provide transaction anomaly scoring and compliance checks that
// feed into the broader ledger and consensus systems.
type ForensicNodeInterface interface {
	NodeInterface
	AnalyseTransaction(tx []byte) (float32, error)
	ComplianceCheck(tx []byte, threshold float32) (float32, error)
	StartMonitoring(ctx context.Context, txCh <-chan []byte, threshold float32)
// CustodialNodeInterface exposes asset custody operations.
type CustodialNodeInterface interface {
	NodeInterface
	Register(addr string) error
	Deposit(addr, token string, amount uint64) error
	Withdraw(addr, token string, amount uint64) error
	Transfer(from, to, token string, amount uint64) error
	BalanceOf(addr, token string) (uint64, error)
	Audit() ([]byte, error)

// QuantumNodeInterface extends NodeInterface with quantum-safe operations.
type QuantumNodeInterface interface {
	NodeInterface
	SecureBroadcast(topic string, data []byte) error
	SecureSubscribe(topic string) (<-chan []byte, error)
	Sign(msg []byte) ([]byte, error)
	Verify(msg, sig []byte) (bool, error)
	RotateKeys() error
// AIEnhancedNodeInterface extends NodeInterface with AI powered helpers.
// Parameters are kept generic (byte slices) to avoid direct core dependencies
// while still allowing advanced functionality when implemented in the core
// package.
type AIEnhancedNodeInterface interface {
	NodeInterface

	// PredictLoad returns the predicted transaction volume for the provided
	// metrics blob. The caller defines the encoding of the blob.
	PredictLoad([]byte) (uint64, error)

	// AnalyseTx performs batch anomaly detection over the provided
	// transaction list. Keys in the returned map are hex-encoded hashes.
	AnalyseTx([]byte) (map[string]float32, error)
// EnergyNodeInterface extends NodeInterface with energy tracking methods.
type EnergyNodeInterface interface {
	NodeInterface
	RecordUsage(txs uint64, kwh float64) error
	Efficiency() (float64, error)
	NetworkAverage() (float64, error)
// IntegrationNodeInterface extends NodeInterface with integration specific
// management helpers. It deliberately avoids referencing core types to keep the
// package dependency hierarchy simple.
type IntegrationNodeInterface interface {
	NodeInterface
	RegisterAPI(name, endpoint string) error
	RemoveAPI(name string) error
	ListAPIs() []string
	ConnectChain(id, endpoint string) error
	DisconnectChain(id string) error
	ListChains() []string
// RegulatoryNodeInterface extends NodeInterface with compliance helpers.
type RegulatoryNodeInterface interface {
	NodeInterface
	VerifyTransaction([]byte) error
	VerifyKYC([]byte) error
	EraseKYC(string) error
	RiskScore(string) int
	GenerateReport() ([]byte, error)
// DisasterRecovery interface extends NodeInterface with backup and restore
// helpers used by specialised disaster recovery nodes. Implementations may
// persist snapshots to multiple locations and verify integrity before applying
// them to the ledger.
type DisasterRecovery interface {
	NodeInterface
	Start()
	Stop() error
	BackupNow(ctx context.Context, incremental bool) error
	Restore(path string) error
	Verify(path string) error

// ContentMeta describes stored content pinned by a content node.
type ContentMeta struct {
	CID      string
	Size     uint64
	Uploaded time.Time
}

// ContentNodeInterface extends NodeInterface with large content operations.
type ContentNodeInterface interface {
	NodeInterface
	StoreContent(data, key []byte) (string, error)
	RetrieveContent(cid string, key []byte) ([]byte, error)
	ListContent() ([]ContentMeta, error)

// ZKPNodeInterface extends NodeInterface with zero-knowledge proof functions.
type ZKPNodeInterface interface {
	NodeInterface
	GenerateProof(data []byte) ([]byte, error)
	VerifyProof(data, proof []byte) bool
	StoreProof(txID string, proof []byte)
	Proof(txID string) ([]byte, bool)
	SubmitTransaction(tx any, proof []byte) error
// Address mirrors the core.Address type without importing the core package.
type Address [20]byte

// LedgerAuditEvent mirrors the core ledger audit event structure.
type LedgerAuditEvent struct {
	Timestamp int64
	Address   Address
	Event     string
	Meta      map[string]string
}

// AuditNodeInterface extends NodeInterface with audit management functions.
type AuditNodeInterface interface {
	NodeInterface
	LogAudit(addr Address, event string, meta map[string]string) error
	AuditEvents(addr Address) ([]LedgerAuditEvent, error)


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
