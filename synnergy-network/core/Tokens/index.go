package Tokens


import core "synnergy-network/core"
import "time"

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
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
=======
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
