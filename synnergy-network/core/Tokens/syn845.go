package Tokens

import (
	"fmt"
	"sync"
	"time"
)

type Address [20]byte

// DebtMetadata stores comprehensive data about a debt instrument.
type DebtMetadata struct {
	ID              string
	Borrower        Address
	Principal       uint64
	InterestRate    float64
	PenaltyRate     float64
	IssueDate       time.Time
	DueDate         time.Time
	Currency        string
	Collateral      string
	Status          string
	PaidAmount      uint64
	AccruedInterest uint64
}

// PaymentRecord captures each repayment entry.
type PaymentRecord struct {
	Date      time.Time
	Amount    uint64
	Interest  uint64
	Principal uint64
	Remaining uint64
}

// SYN845Token implements the debt token logic.
type SYN845Token struct {
	*BaseToken
	mu       sync.RWMutex
	Debts    map[string]*DebtMetadata
	Payments map[string][]PaymentRecord
}

// NewSYN845Token constructs a debt token with metadata.
func NewSYN845Token(meta Metadata, init map[Address]uint64) *SYN845Token {
	meta.Standard = StdSYN845
	bt := &BaseToken{id: deriveID(meta.Standard), meta: meta, balances: NewBalanceTable()}
	t := &SYN845Token{BaseToken: bt, Debts: make(map[string]*DebtMetadata), Payments: make(map[string][]PaymentRecord)}
	for a, v := range init {
		t.balances.Set(t.id, a, v)
		t.meta.TotalSupply += v
	}
	RegisterToken(t)
	return t
}

// IssueDebt registers a new debt record.
func (t *SYN845Token) IssueDebt(rec DebtMetadata) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.Debts[rec.ID]; ok {
		return fmt.Errorf("debt exists")
	}
	rec.IssueDate = time.Now().UTC()
	rec.Status = "active"
	t.Debts[rec.ID] = &rec
	return nil
}

// RecordPayment logs a repayment and updates balances.
func (t *SYN845Token) RecordPayment(id string, amount uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	d, ok := t.Debts[id]
	if !ok {
		return fmt.Errorf("debt not found")
	}
	if d.Status != "active" {
		return fmt.Errorf("debt not active")
	}
	interest := uint64(float64(d.Principal-d.PaidAmount) * d.InterestRate / 100)
	if amount > interest {
		principal := amount - interest
		if principal > d.Principal-d.PaidAmount {
			principal = d.Principal - d.PaidAmount
		}
		d.PaidAmount += principal
	}
	d.AccruedInterest += interest
	remaining := d.Principal - d.PaidAmount
	if remaining == 0 {
		d.Status = "repaid"
	}
	t.Payments[id] = append(t.Payments[id], PaymentRecord{Date: time.Now().UTC(), Amount: amount, Interest: interest, Principal: d.PaidAmount, Remaining: remaining})
	return nil
}

// AdjustInterest changes the interest rate for a debt.
func (t *SYN845Token) AdjustInterest(id string, rate float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	d, ok := t.Debts[id]
	if !ok {
		return fmt.Errorf("debt not found")
	}
	d.InterestRate = rate
	return nil
}

// MarkDefault flags the debt as defaulted.
func (t *SYN845Token) MarkDefault(id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	d, ok := t.Debts[id]
	if !ok {
		return fmt.Errorf("debt not found")
	}
	d.Status = "default"
	return nil
}

// DebtInfo returns details for a debt instrument.
func (t *SYN845Token) DebtInfo(id string) (DebtMetadata, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	d, ok := t.Debts[id]
	if !ok {
		return DebtMetadata{}, false
	}
	cp := *d
	return cp, true
}

// ListDebts returns all debt records.
func (t *SYN845Token) ListDebts() []DebtMetadata {
	t.mu.RLock()
	defer t.mu.RUnlock()
	list := make([]DebtMetadata, 0, len(t.Debts))
	for _, d := range t.Debts {
		list = append(list, *d)
	}
	return list
}
