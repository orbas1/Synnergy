package Nodes

import (
	"context"
	"time"
)

// Address mirrors the core address type without creating a dependency.
type Address [20]byte

// Hash mirrors the core hash type.
type Hash [32]byte

// NodeInterface defines minimal node behaviour independent from core types.
type NodeInterface interface {
	DialSeed([]string) error
	Broadcast(topic string, data []byte) error
	Subscribe(topic string) (<-chan []byte, error)
	ListenAndServe()
	Close() error
	Peers() []string
}

// ValidatorNodeInterface extends NodeInterface with validator specific actions.
type ValidatorNodeInterface interface {
	NodeInterface
	EnablePoH(bool)
	EnablePoS(bool)
	EnablePoW(bool)
	Start()
	Stop() error
	ValidateTx([]byte) error
	ProposeBlock() error
	VoteBlock([]byte, []byte) error
}

// FullNodeAPI exposes the extended functionality provided by a Synnergy full
// node.
type FullNodeAPI interface {
	NodeInterface
	Start()
	Stop() error
	Ledger() any
	Mode() uint8
}

// ElectedAuthorityNodeInterface defines privileged actions for elected
// authority nodes.
type ElectedAuthorityNodeInterface interface {
	NodeInterface
	RecordVote(addr Address)
	ReportMisbehaviour(addr Address)
	ValidateTransaction(tx []byte) error
	CreateBlock(blob []byte) error
	ReverseTransaction(hash Hash, sigs [][]byte) error
	ViewPrivateTransaction(hash Hash) ([]byte, error)
	ApproveLoanProposal(id string) error
}

// MiningNodeInterface extends NodeInterface with mining specific controls.
type MiningNodeInterface interface {
	NodeInterface
	StartMining()
	StopMining() error
	AddTransaction(tx []byte) error
}

// MasterNodeInterface defines additional services used by master nodes.
type MasterNodeInterface interface {
	NodeInterface
	ProcessTx(tx any) error
	HandlePrivateTx(tx any, key []byte) error
	VoteProposal(id uint64, approve bool) error
	Start()
	Stop() error
}

// StakingNodeInterface exposes staking-related actions.
type StakingNodeInterface interface {
	NodeInterface
	Stake(addr string, amount uint64) error
	Unstake(addr string, amount uint64) error
	ProposeBlock(data []byte) error
	ValidateBlock(data []byte) error
	Status() string
}

// LightNodeInterface defines the minimal behaviour for light clients.
type LightNodeInterface interface {
	NodeInterface
	StoreHeader(h BlockHeader)
	Headers() []BlockHeader
}

// IndexingNodeInterface describes capabilities of indexing nodes.
type IndexingNodeInterface interface {
	NodeInterface
	AddBlock(b any)
	QueryTxHistory(addr Address) []any
	QueryState(addr Address, key string) (any, bool)
}

// GatewayInterface extends NodeInterface with cross-chain and data functions.
type GatewayInterface interface {
	NodeInterface
	ConnectChain(local, remote string) (any, error)
	DisconnectChain(id string) error
	ListConnections() any
	RegisterExternalSource(name, url string)
	RemoveExternalSource(name string)
	ExternalSources() map[string]string
	PushExternalData(name string, data []byte) error
	QueryExternalData(name string) ([]byte, error)
}

// APINodeInterface exposes HTTP API lifecycle controls.
type APINodeInterface interface {
	NodeInterface
	APINode_Start(addr string) error
	APINode_Stop() error
}

// Watchtower defines the interface implemented by watchtower nodes.
type Watchtower interface {
	NodeInterface
	Alerts() <-chan string
}

// ForensicNodeInterface exposes forensic analysis helpers.
type ForensicNodeInterface interface {
	NodeInterface
	AnalyseTransaction(tx []byte) (float32, error)
	ComplianceCheck(tx []byte, threshold float32) (float32, error)
	StartMonitoring(ctx context.Context, txCh <-chan []byte, threshold float32)
}

// CustodialNodeInterface exposes asset custody operations.
type CustodialNodeInterface interface {
	NodeInterface
	Register(addr string) error
	Deposit(addr, token string, amount uint64) error
	Withdraw(addr, token string, amount uint64) error
	Transfer(from, to, token string, amount uint64) error
	BalanceOf(addr, token string) (uint64, error)
	Audit() ([]byte, error)
}

// QuantumNodeInterface exposes quantum-safe operations.
type QuantumNodeInterface interface {
	NodeInterface
	SecureBroadcast(topic string, data []byte) error
	SecureSubscribe(topic string) (<-chan []byte, error)
	Sign(msg []byte) ([]byte, error)
	Verify(msg, sig []byte) (bool, error)
	RotateKeys() error
}

// AIEnhancedNodeInterface defines AI-powered helpers.
type AIEnhancedNodeInterface interface {
	NodeInterface
	PredictLoad([]byte) (uint64, error)
	AnalyseTx([]byte) (map[string]float32, error)
}

// EnergyNodeInterface exposes energy tracking methods.
type EnergyNodeInterface interface {
	NodeInterface
	RecordUsage(txs uint64, kwh float64) error
	Efficiency() (float64, error)
	NetworkAverage() (float64, error)
}

// IntegrationNodeInterface exposes integration management helpers.
type IntegrationNodeInterface interface {
	NodeInterface
	RegisterAPI(name, endpoint string) error
	RemoveAPI(name string) error
	ListAPIs() []string
	ConnectChain(id, endpoint string) error
	DisconnectChain(id string) error
	ListChains() []string
}

// RegulatoryNodeInterface exposes compliance helpers.
type RegulatoryNodeInterface interface {
	NodeInterface
	VerifyTransaction([]byte) error
	VerifyKYC([]byte) error
	EraseKYC(string) error
	RiskScore(string) int
	GenerateReport() ([]byte, error)
}

// DisasterRecovery exposes backup and restore helpers.
type DisasterRecovery interface {
	NodeInterface
	Start()
	Stop() error
	BackupNow(ctx context.Context, incremental bool) error
	Restore(path string) error
	Verify(path string) error
}

// ContentMeta describes stored content pinned by a content node.
type ContentMeta struct {
	CID      string
	Size     uint64
	Uploaded time.Time
}

// ContentNodeInterface exposes large content operations.
type ContentNodeInterface interface {
	NodeInterface
	StoreContent(data, key []byte) (string, error)
	RetrieveContent(cid string, key []byte) ([]byte, error)
	ListContent() ([]ContentMeta, error)
}

// ZKPNodeInterface exposes zero-knowledge proof functions.
type ZKPNodeInterface interface {
	NodeInterface
	GenerateProof(data []byte) ([]byte, error)
	VerifyProof(data, proof []byte) bool
	StoreProof(txID string, proof []byte)
	Proof(txID string) ([]byte, bool)
	SubmitTransaction(tx any, proof []byte) error
}

// LedgerAuditEvent mirrors the core ledger audit event structure.
type LedgerAuditEvent struct {
	Timestamp int64
	Address   Address
	Event     string
	Meta      map[string]string
}

// AuditNodeInterface exposes audit management functions.
type AuditNodeInterface interface {
	NodeInterface
	LogAudit(addr Address, event string, meta map[string]string) error
	AuditEvents(addr Address) ([]LedgerAuditEvent, error)
}

// AutonomousAgent defines additional behaviour for autonomous nodes.
type AutonomousAgent interface {
	NodeInterface
	AddRule(rule any)
	RemoveRule(id string)
	Start()
	Stop() error
}

// HolographicNodeInterface exposes holographic functions.
type HolographicNodeInterface interface {
	NodeInterface
	EncodeStore(data []byte) (any, error)
	Retrieve(id any) ([]byte, error)
	SyncConsensus(c any) error
	ProcessTx(tx any) error
	ExecuteContract(ctx any, vm any, code []byte) error
}

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
}

// EnvironmentalMonitoringInterface exposes sensor management helpers.
type EnvironmentalMonitoringInterface interface {
	NodeInterface
	RegisterSensor(id, endpoint string) error
	RemoveSensor(id string) error
	ListSensors() ([]string, error)
	AddTrigger(id string, threshold float64, action string) error
	Start()
	Stop() error
}

// BiometricSecurityNode extends NodeInterface with biometric operations.
type BiometricSecurityNode interface {
	NodeInterface
	Enroll(addr string, data []byte) error
	Verify(addr string, data []byte) bool
	Delete(addr string)
	ValidateTransaction(tx any, data []byte) bool
}

// BankInstitutionalNode defines behaviour for specialised bank nodes.
type BankInstitutionalNode interface {
	NodeInterface
	MonitorTransaction(data []byte) error
	ComplianceReport() ([]byte, error)
	ConnectFinancialNetwork(endpoint string) error
	UpdateRuleset(rules map[string]any)
}

// WarfareNodeInterface is implemented by nodes specialised for military operations.
type WarfareNodeInterface interface {
	NodeInterface
	SecureCommand(data []byte) error
	TrackLogistics(itemID, status string) error
	ShareTactical(data []byte) error
}

// MobileMiner extends NodeInterface with light mining controls.
type MobileMiner interface {
	NodeInterface
	StartMining()
	StopMining()
	SetIntensity(int)
	Stats() any
}

// CentralBankingNode defines the extended behaviour required by central bank infrastructure.
type CentralBankingNode interface {
	NodeInterface
	SetInterestRate(float64) error
	InterestRate() float64
	SetReserveRequirement(float64) error
	ReserveRequirement() float64
	IssueDigitalCurrency(addr [20]byte, amount uint64) error
	RecordSettlement(tx []byte) error
}
