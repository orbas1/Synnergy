package Tokens

// index.go – collection of lightweight interfaces and structs that expose token
// metadata to packages outside of the core without introducing circular
// dependencies.  Each specialised token lives in its own file within this
// directory; this file simply aggregates the common type declarations.

import "time"

// Address mirrors core.Address but is redefined here to keep this package free
// of direct core dependencies.
type Address [20]byte

// AddressZero is the zero-value address (all bytes zero).
var AddressZero Address

// TokenStandard mirrors core.TokenStandard.  It enumerates every supported token
// specification within the Synnergy ecosystem.  Each constant corresponds to a
// dedicated file containing the token's implementation (e.g. syn10.go,
// syn845.go, etc.).
type TokenStandard uint16

const (
	StdSYN10   TokenStandard = 10
	StdSYN11   TokenStandard = 11
	StdSYN12   TokenStandard = 12
	StdSYN20   TokenStandard = 20
	StdSYN70   TokenStandard = 70
	StdSYN130  TokenStandard = 130
	StdSYN131  TokenStandard = 131
	StdSYN200  TokenStandard = 200
	StdSYN223  TokenStandard = 223
	StdSYN300  TokenStandard = 300
	StdSYN500  TokenStandard = 500
	StdSYN600  TokenStandard = 600
	StdSYN700  TokenStandard = 700
	StdSYN721  TokenStandard = 721
	StdSYN722  TokenStandard = 722
	StdSYN800  TokenStandard = 800
	StdSYN845  TokenStandard = 845
	StdSYN900  TokenStandard = 900
	StdSYN1000 TokenStandard = 1000
	StdSYN1100 TokenStandard = 1100
	StdSYN1155 TokenStandard = 1155
	StdSYN1200 TokenStandard = 1200
	StdSYN1300 TokenStandard = 1300
	StdSYN1401 TokenStandard = 1401
	StdSYN1500 TokenStandard = 1500
	StdSYN1600 TokenStandard = 1600
	StdSYN1700 TokenStandard = 1700
	StdSYN1800 TokenStandard = 1800
	StdSYN1900 TokenStandard = 1900
	StdSYN1967 TokenStandard = 1967
	StdSYN2100 TokenStandard = 2100
	StdSYN2200 TokenStandard = 2200
	StdSYN2369 TokenStandard = 2369
	StdSYN2400 TokenStandard = 2400
	StdSYN2500 TokenStandard = 2500
	StdSYN2600 TokenStandard = 2600
	StdSYN2700 TokenStandard = 2700
	StdSYN2800 TokenStandard = 2800
	StdSYN2900 TokenStandard = 2900
	StdSYN3000 TokenStandard = 3000
	StdSYN3100 TokenStandard = 3100
	StdSYN3200 TokenStandard = 3200
	StdSYN3300 TokenStandard = 3300
	StdSYN3400 TokenStandard = 3400
	StdSYN3500 TokenStandard = 3500
	StdSYN3600 TokenStandard = 3600
	StdSYN3700 TokenStandard = 3700
	StdSYN3800 TokenStandard = 3800
	StdSYN3900 TokenStandard = 3900
	StdSYN4200 TokenStandard = 4200
	StdSYN4300 TokenStandard = 4300
	StdSYN4700 TokenStandard = 4700
	StdSYN4900 TokenStandard = 4900
	StdSYN5000 TokenStandard = 5000
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
