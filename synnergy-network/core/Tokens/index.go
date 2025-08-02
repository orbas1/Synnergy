package Tokens

// index.go – collection of lightweight interfaces and structs that expose token
// metadata to packages outside of the core without introducing circular
// dependencies.  Each specialised token lives in its own file within this
// directory; this file simply aggregates the common type declarations.

import "time"

// Address mirrors core.Address but is redefined here to keep this package free
// of direct core dependencies.
type Address [20]byte

// TokenStandard mirrors core.TokenStandard.  It enumerates every supported token
// specification within the Synnergy ecosystem.  Each constant corresponds to a
// dedicated file containing the token's implementation (e.g. syn10.go,
// syn845.go, etc.).
type TokenStandard uint16

const (
	StdSYN10   TokenStandard = 1
	StdSYN11   TokenStandard = 11
	StdSYN12   TokenStandard = 12
	StdSYN20   TokenStandard = 2
	StdSYN70   TokenStandard = 7
	StdSYN130  TokenStandard = 13
	StdSYN131  TokenStandard = 131
	StdSYN200  TokenStandard = 20
	StdSYN223  TokenStandard = 22
	StdSYN300  TokenStandard = 30
	StdSYN500  TokenStandard = 50
	StdSYN600  TokenStandard = 60
	StdSYN700  TokenStandard = 70
	StdSYN721  TokenStandard = 72
	StdSYN722  TokenStandard = 72
	StdSYN800  TokenStandard = 80
	StdSYN845  TokenStandard = 84
	StdSYN900  TokenStandard = 90
	StdSYN1000 TokenStandard = 100
	StdSYN1100 TokenStandard = 110
	StdSYN1155 TokenStandard = 115
	StdSYN1200 TokenStandard = 120
	StdSYN1300 TokenStandard = 130
	StdSYN1401 TokenStandard = 140
	StdSYN1500 TokenStandard = 150
	StdSYN1600 TokenStandard = 160
	StdSYN1700 TokenStandard = 170
	StdSYN1800 TokenStandard = 180
	StdSYN1900 TokenStandard = 190
	StdSYN1967 TokenStandard = 196
	StdSYN2100 TokenStandard = 210
	StdSYN2200 TokenStandard = 220
	StdSYN2369 TokenStandard = 236
	StdSYN2400 TokenStandard = 240
	StdSYN2500 TokenStandard = 250
	StdSYN2600 TokenStandard = 260
	StdSYN2700 TokenStandard = 270
	StdSYN2800 TokenStandard = 280
	StdSYN2900 TokenStandard = 290
	StdSYN3000 TokenStandard = 300
	StdSYN3100 TokenStandard = 310
	StdSYN3200 TokenStandard = 320
	StdSYN3300 TokenStandard = 330
	StdSYN3400 TokenStandard = 340
	StdSYN3500 TokenStandard = 350
	StdSYN3600 TokenStandard = 360
	StdSYN3700 TokenStandard = 370
	StdSYN3800 TokenStandard = 380
	StdSYN3900 TokenStandard = 390
	StdSYN4200 TokenStandard = 420
	StdSYN4300 TokenStandard = 430
	StdSYN4700 TokenStandard = 470
	StdSYN4900 TokenStandard = 490
	StdSYN5000 TokenStandard = 500
)

// -----------------------------------------------------------------------------
// Generic token interfaces
// -----------------------------------------------------------------------------

// TokenInterfaces represents the minimal behaviour shared by all tokens.  Many
// interfaces in this package embed TokenInterfaces to advertise that they expose
// basic metadata via the Meta method.
type TokenInterfaces interface {
	Meta() any
}

// -----------------------------------------------------------------------------
// Carbon footprint (SYN1800)
// -----------------------------------------------------------------------------

// CarbonFootprintRecord represents a carbon footprint event recorded on-chain.
type CarbonFootprintRecord struct {
	ID          uint64
	Owner       Address
	Amount      int64
	Issued      int64
	Description string
	Source      string
}

// CarbonFootprintTokenAPI defines the external interface for SYN1800 tokens.
type CarbonFootprintTokenAPI interface {
	TokenInterfaces
	RecordEmission(owner Address, amt int64, desc, src string) (uint64, error)
	RecordOffset(owner Address, amt int64, desc, src string) (uint64, error)
	NetBalance(owner Address) int64
	ListRecords(owner Address) ([]CarbonFootprintRecord, error)
}

// -----------------------------------------------------------------------------
// SYN500 utility token
// -----------------------------------------------------------------------------

type AccessInfo struct {
	Tier         uint8
	MaxUsage     uint64
	UsageCount   uint64
	Expiry       int64
	RewardPoints uint64
}

type SYN500 interface {
	TokenInterfaces
	GrantAccess(addr Address, tier uint8, max uint64, expiry int64)
	UpdateAccess(addr Address, tier uint8, max uint64, expiry int64)
	RevokeAccess(addr Address)
	RecordUsage(addr Address, points uint64) error
	RedeemReward(addr Address, points uint64) error
	RewardBalance(addr Address) uint64
	Usage(addr Address) uint64
	AccessInfoOf(addr Address) (AccessInfo, bool)
}

// -----------------------------------------------------------------------------
// Employment tokens (SYN3100)
// -----------------------------------------------------------------------------

type EmploymentContractMeta struct {
	ContractID string
	Employer   Address
	Employee   Address
	Position   string
	Salary     uint64
	Benefits   string
	Start      int64
	End        int64
	Active     bool
}

type EmploymentToken interface {
	TokenInterfaces
	CreateContract(EmploymentContractMeta) error
	PaySalary(id string) error
	UpdateBenefits(id, benefits string) error
	TerminateContract(id string) error
	GetContract(id string) (EmploymentContractMeta, bool)
}

// -----------------------------------------------------------------------------
// Financial token helpers
// -----------------------------------------------------------------------------

type ForexToken interface {
	TokenInterfaces
	Rate() float64
	Pair() string
}

type CurrencyToken interface {
	TokenInterfaces
	UpdateRate(float64)
	Info() (string, string, float64, time.Time)
	MintCurrency(Address, uint64) error
	RedeemCurrency(Address, uint64) error
}

// Futures tokens (SYN3600)
type FuturesTokenInterface interface {
	TokenInterfaces
	UpdatePrice(uint64)
	OpenPosition(addr Address, size, entryPrice uint64, long bool, margin uint64) error
	ClosePosition(addr Address, exitPrice uint64) (int64, error)
}

// Legal agreements (SYN4700)
type LegalTokenAPI interface {
	TokenInterfaces
	AddSignature(party any, sig []byte)
	RevokeSignature(party any)
	UpdateStatus(status string)
	StartDispute()
	ResolveDispute(result string)
}

// Reward tokens (SYN600)
type RewardTokenInterface interface {
	TokenInterfaces
	Stake(addr any, amount uint64, duration int64) error
	Unstake(addr any) error
	AddEngagement(addr any, points uint64) error
	EngagementOf(addr any) uint64
	DistributeStakingRewards(rate uint64) error
}

// Asset backed tokens (SYN800)
type AssetMeta struct {
	Description string
	Valuation   uint64
	Location    string
	AssetType   string
	Certified   bool
	Compliance  string
	Updated     time.Time
}

type SYN800 interface {
	TokenInterfaces
	RegisterAsset(meta AssetMeta) error
	UpdateValuation(val uint64) error
	GetAsset() (AssetMeta, error)
}

// Intellectual property tokens (SYN700)
type SYN700 interface {
	TokenInterfaces
	RegisterIPAsset(id string, meta any, owner any) error
	TransferIPOwnership(id string, from, to any, share uint64) error
	CreateLicense(id string, license any) error
	RevokeLicense(id string, licensee any) error
	RecordRoyalty(id string, licensee any, amount uint64) error
}

// Supply chain tokens (SYN1300)
type SupplyChainAsset struct {
	ID          string
	Description string
	Location    string
	Status      string
	Owner       Address
	Timestamp   time.Time
}

type SupplyChainEvent struct {
	AssetID     string
	Description string
	Location    string
	Status      string
	Owner       Address
	Timestamp   time.Time
}

type SupplyChainToken interface {
	TokenInterfaces
	RegisterAsset(SupplyChainAsset) error
	RecordEvent(SupplyChainEvent) error
	AssetHistory(id string) []SupplyChainEvent
}

// -----------------------------------------------------------------------------
// Education credits (SYN1900)
// -----------------------------------------------------------------------------

type EducationCreditMetadata struct {
	CreditID    string
	Recipient   string
	Institution string
	Credits     uint64
	IssueDate   time.Time
}

type EducationCreditTokenInterface interface {
	TokenInterfaces
	IssueCredit(EducationCreditMetadata) error
	GetCredit(id string) (EducationCreditMetadata, bool)
	ListCredits(recipient string) []EducationCreditMetadata
	VerifyCredit(id string) bool
	RevokeCredit(id string) error
}

// End of file
