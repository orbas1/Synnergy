package Tokens



// CarbonFootprintRecord represents a carbon footprint event recorded on-chain.
type CarbonFootprintRecord struct {
	ID          uint64
	Owner       [20]byte
	Amount      int64
	Issued      int64
	Description string
	Source      string
}

// CarbonFootprintTokenAPI defines the external interface for SYN1800 tokens.
type CarbonFootprintTokenAPI interface {
	RecordEmission(owner [20]byte, amt int64, desc, src string) (uint64, error)
	RecordOffset(owner [20]byte, amt int64, desc, src string) (uint64, error)
	NetBalance(owner [20]byte) int64
	ListRecords(owner [20]byte) ([]CarbonFootprintRecord, error)
}

import core "synnergy-network/core"
import "time"

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// AssetMeta mirrors the metadata used by SYN800 asset-backed tokens.
type AssetMeta struct {
	Description string
	Valuation   uint64
	Location    string
	AssetType   string
	Certified   bool
	Compliance  string
	Updated     time.Time
}

// SYN800 defines the exported interface for asset-backed tokens.
type SYN800 interface {
	TokenInterfaces
	RegisterAsset(meta AssetMeta) error
	UpdateValuation(val uint64) error
	GetAsset() (AssetMeta, error)
}
=======
// SYN700 defines the minimal methods for the intellectual property token
// standard without referencing core types.
type SYN700 interface {
	TokenInterfaces
	RegisterIPAsset(id string, meta any, owner any) error
	TransferIPOwnership(id string, from, to any, share uint64) error
	CreateLicense(id string, license any) error
	RevokeLicense(id string, licensee any) error
	RecordRoyalty(id string, licensee any, amount uint64) error
}
// Address mirrors core.Address to avoid circular dependency.
type Address [20]byte

// SupplyChainAsset describes an asset tracked by SYN1300 tokens.
type SupplyChainAsset struct {
	ID          string
	Description string
	Location    string
	Status      string
	Owner       Address
	Timestamp   time.Time
}

// SupplyChainEvent details movements or updates to an asset.
type SupplyChainEvent struct {
	AssetID     string
	Description string
	Location    string
	Status      string
	Timestamp   time.Time
}

// SupplyChainToken exposes the SYN1300 interface.
type SupplyChainToken interface {
	TokenInterfaces
	RegisterAsset(SupplyChainAsset) error
	UpdateLocation(id, location string) error
	UpdateStatus(id, status string) error
	TransferAsset(id string, newOwner Address) error
	Asset(id string) (SupplyChainAsset, bool)
	Events(id string) []SupplyChainEvent
}
// SYN1600 defines the behaviour expected from music royalty tokens.
type SYN1600 interface {
	TokenInterfaces
	AddRevenue(amount uint64, txID string)
	RevenueHistory() []any
	DistributeRoyalties(amount uint64) error
	UpdateInfo(info any)
}
// EducationCreditMetadata captures the details of a single education credit.
type EducationCreditMetadata struct {
	CreditID       string
	CourseID       string
	CourseName     string
	Issuer         string
	Recipient      string
	CreditValue    uint32
	IssueDate      time.Time
	ExpirationDate time.Time
	Metadata       string
	Signature      []byte
}

// CourseRecord stores information about an education course.
type CourseRecord struct {
	ID          string
	Name        string
	Description string
	CreditValue uint32
}

// IssuerRecord stores issuer metadata.
type IssuerRecord struct {
	ID   string
	Name string
	Info string
}

// RecipientRecord stores recipient metadata.
type RecipientRecord struct {
	ID   string
	Name string
}

// VerificationLog records verification and revocation events.
type VerificationLog struct {
	Timestamp time.Time
	CreditID  string
	Verifier  string
	Action    string
	Valid     bool
}

// EducationCreditTokenInterface defines functions unique to SYN1900 tokens.
type EducationCreditTokenInterface interface {
	TokenInterfaces
	IssueCredit(EducationCreditMetadata) error
	VerifyCredit(string) bool
	RevokeCredit(string) error
	GetCredit(string) (EducationCreditMetadata, bool)
	ListCredits(string) []EducationCreditMetadata
}
// Address is a 20 byte account identifier used for cross-package compatibility.
type Address [20]byte

// FinancialDocument mirrors the core representation used by the SYN2100 token
// standard. It avoids a dependency on the main core package.
type FinancialDocument struct {
	DocumentID   string
	DocumentType string
	Issuer       Address
	Recipient    Address
	Amount       uint64
	IssueDate    int64 // unix seconds
	DueDate      int64 // unix seconds
	Description  string
	Financed     bool
	AuditTrail   []string
}

// SupplyFinance exposes the SYN2100 token functionality without pulling in the
// heavy core dependencies.
type SupplyFinance interface {
	TokenInterfaces
	RegisterDocument(FinancialDocument) error
	FinanceDocument(id string, financier Address) error
	GetDocument(id string) (FinancialDocument, bool)
	ListDocuments() []FinancialDocument
	AddLiquidity(addr Address, amount uint64) error
	RemoveLiquidity(addr Address, amount uint64) error
	LiquidityOf(addr Address) uint64
}
// RealTimePayments defines the SYN2200 payment functions.
type RealTimePayments interface {
	SendPayment(from, to core.Address, amount uint64, currency string) (uint64, error)
	Payment(id uint64) (PaymentRecord, bool)
}

// NewSYN2200 exposes constructor for external packages.
func NewSYN2200(meta core.Metadata, init map[core.Address]uint64, ledger *core.Ledger, gas core.GasCalculator) (*SYN2200Token, error) {
	return NewSYN2200Token(meta, init, ledger, gas)
}
// DataMarketplace defines behaviour for SYN2400 tokens.
type DataMarketplace interface {
	TokenInterfaces
	UpdateMetadata(hash, desc string)
	SetPrice(p uint64)
	SetStatus(s string)
	GrantAccess(addr [20]byte, rights string)
	RevokeAccess(addr [20]byte)
	HasAccess(addr [20]byte) bool
}

// Address mirrors core.Address without pulling the full dependency.
type Address [20]byte

// Syn2500Member records DAO membership details.
type Syn2500Member struct {
	DAOID       string    `json:"dao_id"`
	Address     Address   `json:"address"`
	VotingPower uint64    `json:"voting_power"`
	Issued      time.Time `json:"issued"`
	Active      bool      `json:"active"`
	Delegate    Address   `json:"delegate"`
}

// Syn2500Token defines the external interface for DAO tokens.
type Syn2500Token struct {
	Members map[Address]Syn2500Member
}

func (t *Syn2500Token) Meta() any { return t }
