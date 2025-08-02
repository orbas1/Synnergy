package core

import (
	"sync"
	"time"
)

// Bill represents a bill record managed by the SYN3200 standard.
type Bill struct {
	ID             uint64
	Issuer         Address
	Payer          Address
	OriginalAmount uint64
	Remaining      uint64
	DueDate        time.Time
	Paid           bool
	Meta           string
	Payments       []Payment
	Adjustments    []Adjustment
}

type Payment struct {
	Amount uint64
	Date   time.Time
}

type Adjustment struct {
	NewAmount uint64
	Date      time.Time
}

// Syn3200Token extends core.BaseToken with bill management features.
type Syn3200Token struct {
	*BaseToken
	ledger *Ledger
	gas    GasCalculator
	mu     sync.RWMutex
	bills  map[uint64]*Bill
	nextID uint64
}

// NewSyn3200 creates a new SYN3200 token standard bound to the given ledger.
func NewSyn3200(meta Metadata, ledger *Ledger, gas GasCalculator) *Syn3200Token {
	tok, _ := (Factory{}).Create(meta, map[Address]uint64{})
	bt := tok.(*BaseToken)
	bt.ledger = ledger
	bt.gas = gas
	t := &Syn3200Token{BaseToken: bt, ledger: ledger, gas: gas, bills: make(map[uint64]*Bill)}
	RegisterToken(t)
	return t
}

// CreateBill records a new bill and mints tokens equal to the amount.
func (t *Syn3200Token) CreateBill(issuer, payer Address, amount uint64, due time.Time, meta string) uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	id := t.nextID
	t.nextID++
	b := &Bill{ID: id, Issuer: issuer, Payer: payer, OriginalAmount: amount, Remaining: amount, DueDate: due, Meta: meta}
	t.bills[id] = b
	_ = t.Mint(payer, amount)
	if t.ledger != nil {
		t.ledger.EmitTransfer(t.ID(), Address{}, payer, amount)
	}
	return id
}

// PayFraction settles part of the bill from payer to issuer.
func (t *Syn3200Token) PayFraction(billID uint64, payer Address, amount uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	b, ok := t.bills[billID]
	if !ok {
		return errInvalidAsset
	}
	if b.Paid || b.Remaining < amount {
		return errInvalidAsset
	}
	if err := t.Transfer(payer, b.Issuer, amount); err != nil {
		return err
	}
	b.Remaining -= amount
	b.Payments = append(b.Payments, Payment{Amount: amount, Date: time.Now().UTC()})
	if b.Remaining == 0 {
		b.Paid = true
	}
	if t.ledger != nil {
		fee := t.gas.Calculate("Syn3200_PayFraction", amount)
		t.ledger.DeductGas(payer, fee)
		t.ledger.EmitTransfer(t.ID(), payer, b.Issuer, amount)
	}
	return nil
}

// AdjustAmount updates the remaining amount of a bill.
func (t *Syn3200Token) AdjustAmount(billID uint64, newAmount uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	b, ok := t.bills[billID]
	if !ok {
		return errInvalidAsset
	}
	b.Remaining = newAmount
	b.Adjustments = append(b.Adjustments, Adjustment{NewAmount: newAmount, Date: time.Now().UTC()})
	return nil
}

// GetBill returns the bill record if present.
func (t *Syn3200Token) GetBill(id uint64) (*Bill, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	b, ok := t.bills[id]
	return b, ok
}

// MetaInterface exposes bill metadata for indexing without core deps.
type MetaInterface interface {
	Meta() Metadata
}

var _ MetaInterface = (*Syn3200Token)(nil)
