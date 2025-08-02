//go:build tokens
// +build tokens

package Tokens

import (
	"sync"
	"time"
)

// PaymentRecord tracks a real-time payment lifecycle.
type RTPaymentRecord struct {
	ID        uint64
	Amount    uint64
	Currency  string
	Sender    Address
	Recipient Address
	Created   time.Time
	Settled   bool
	SettledAt time.Time
}

// SYN2200Token implements the real-time payments token standard.
type SYN2200Token struct {
	*BaseToken
	payments map[uint64]RTPaymentRecord
	payMu    sync.RWMutex
	nextID   uint64
}

// NewSYN2200Token creates and registers a SYN2200 token.
func NewSYN2200Token(meta Metadata, init map[Address]uint64) *SYN2200Token {
	if meta.Created.IsZero() {
		meta.Created = time.Now().UTC()
	}
	bt := &BaseToken{id: deriveID(meta.Standard), meta: meta, balances: NewBalanceTable()}
	for a, v := range init {
		bt.balances.Set(bt.id, a, v)
		bt.meta.TotalSupply += v
	}
	rtp := &SYN2200Token{
		BaseToken: bt,
		payments:  make(map[uint64]RTPaymentRecord),
	}
	RegisterToken(rtp)
	return rtp
}

// SendPayment transfers funds instantly and records the payment.
func (t *SYN2200Token) SendPayment(from, to Address, amount uint64, currency string) (uint64, error) {
	if err := t.Transfer(from, to, amount); err != nil {
		return 0, err
	}
	t.payMu.Lock()
	defer t.payMu.Unlock()
	t.nextID++
	id := t.nextID
	t.payments[id] = RTPaymentRecord{
		ID:        id,
		Amount:    amount,
		Currency:  currency,
		Sender:    from,
		Recipient: to,
		Created:   time.Now().UTC(),
		Settled:   true,
		SettledAt: time.Now().UTC(),
	}
	return id, nil
}

// Payment returns a payment record by ID.
func (t *SYN2200Token) Payment(id uint64) (RTPaymentRecord, bool) {
	t.payMu.RLock()
	defer t.payMu.RUnlock()
	p, ok := t.payments[id]
	return p, ok
}
