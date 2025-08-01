package core

import (
	"time"
)

// SYN11Token represents Central Bank Digital Gilts.
// It extends BaseToken with gilt-specific metadata and behaviour.
type SYN11Token struct {
	BaseToken

	GiltCode   string
	Issuer     string
	Maturity   time.Time
	CouponRate float64
}

// NewSYN11Token creates and registers a new SYN11 token using the factory.
// meta.Standard should be StdSYN11.
func NewSYN11Token(meta Metadata, issuer, giltCode string, maturity time.Time, coupon float64, init map[Address]uint64) (*SYN11Token, error) {
	if meta.Standard == 0 {
		meta.Standard = StdSYN11
	}
	tok, err := (Factory{}).Create(meta, init)
	if err != nil {
		return nil, err
	}
	bt := tok.(*BaseToken)
	t := &SYN11Token{
		BaseToken:  *bt,
		GiltCode:   giltCode,
		Issuer:     issuer,
		Maturity:   maturity,
		CouponRate: coupon,
	}
	RegisterToken(t)
	return t, nil
}

// Issue mints new gilts to the recipient.
func (t *SYN11Token) Issue(to Address, amount uint64) error {
	return t.Mint(to, amount)
}

// Redeem burns gilts from the holder.
func (t *SYN11Token) Redeem(from Address, amount uint64) error {
	return t.Burn(from, amount)
}

// UpdateCoupon sets a new coupon rate.
func (t *SYN11Token) UpdateCoupon(rate float64) {
	t.CouponRate = rate
}

// PayCoupon distributes interest to all holders based on the current coupon rate.
func (t *SYN11Token) PayCoupon() map[Address]uint64 {
	payments := make(map[Address]uint64)
	if t.balances == nil {
		return payments
	}
	t.balances.mu.RLock()
	defer t.balances.mu.RUnlock()
	for addr, bal := range t.balances.balances[t.id] {
		interest := uint64(float64(bal) * t.CouponRate)
		t.balances.balances[t.id][addr] += interest
		payments[addr] = interest
	}
	return payments
}
