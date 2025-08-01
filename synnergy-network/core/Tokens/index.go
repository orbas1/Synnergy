package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
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
