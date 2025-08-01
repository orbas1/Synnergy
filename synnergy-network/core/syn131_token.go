package core

import "time"

// ValuationRecord captures historical valuations for SYN131 assets.
type ValuationRecord struct {
	Value     uint64
	Timestamp time.Time
}

// SaleRecord captures sale transactions for SYN131 assets.
type SaleRecord struct {
	Price     uint64
	Buyer     Address
	Seller    Address
	Timestamp time.Time
}

// RentalAgreement models rental terms for an asset.
type RentalAgreement struct {
	Renter Address
	Start  time.Time
	End    time.Time
	Fee    uint64
}

// LicenseRecord models a licensing agreement for an asset.
type LicenseRecord struct {
	Licensee   Address
	Terms      string
	ValidUntil time.Time
}

// SYN131Token provides advanced management of intangible assets.
type SYN131Token struct {
	*BaseToken
	Values   []ValuationRecord
	Sales    []SaleRecord
	Rentals  []RentalAgreement
	Licenses []LicenseRecord
	Shares   map[Address]uint64
}

// NewSYN131Token creates a new SYN131 token instance.
func NewSYN131Token(meta Metadata, init map[Address]uint64) *SYN131Token {
	bt := &BaseToken{id: deriveID(meta.Standard), meta: meta, balances: NewBalanceTable()}
	for a, v := range init {
		bt.balances.Set(bt.id, a, v)
		bt.meta.TotalSupply += v
	}
	return &SYN131Token{BaseToken: bt, Shares: make(map[Address]uint64)}
}

// UpdateValuation records a new valuation entry.
func (t *SYN131Token) UpdateValuation(value uint64) {
	t.Values = append(t.Values, ValuationRecord{Value: value, Timestamp: time.Now().UTC()})
}

// RecordSale appends a sale to the history log.
func (t *SYN131Token) RecordSale(price uint64, buyer, seller Address) {
	t.Sales = append(t.Sales, SaleRecord{Price: price, Buyer: buyer, Seller: seller, Timestamp: time.Now().UTC()})
}

// AddRental registers a new rental agreement.
func (t *SYN131Token) AddRental(r RentalAgreement) {
	t.Rentals = append(t.Rentals, r)
}

// IssueLicense records a licensing agreement.
func (t *SYN131Token) IssueLicense(l LicenseRecord) {
	t.Licenses = append(t.Licenses, l)
}

// TransferShare moves fractional ownership between parties.
func (t *SYN131Token) TransferShare(from, to Address, share uint64) {
	if t.Shares == nil {
		t.Shares = make(map[Address]uint64)
	}
	if t.Shares[from] < share {
		return
	}
	t.Shares[from] -= share
	t.Shares[to] += share
}

// --- VM opcode wrappers ---
func SYN131_UpdateValuation(t *SYN131Token, value uint64) { t.UpdateValuation(value) }
func SYN131_RecordSale(t *SYN131Token, price uint64, buyer, seller Address) {
	t.RecordSale(price, buyer, seller)
}
func SYN131_AddRental(t *SYN131Token, r RentalAgreement)  { t.AddRental(r) }
func SYN131_IssueLicense(t *SYN131Token, l LicenseRecord) { t.IssueLicense(l) }
func SYN131_TransferShare(t *SYN131Token, from, to Address, share uint64) {
	t.TransferShare(from, to, share)
}
