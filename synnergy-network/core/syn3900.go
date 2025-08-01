package core

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// BenefitRecord holds metadata for a government benefit token issuance.
type BenefitRecord struct {
	ID         Hash    `json:"id"`
	Recipient  Address `json:"recipient"`
	Amount     uint64  `json:"amount"`
	ValidFrom  int64   `json:"valid_from"`
	ValidUntil int64   `json:"valid_until"`
	IssuedAt   int64   `json:"issued_at"`
	Conditions string  `json:"conditions"`
	Claimed    bool    `json:"claimed"`
}

// BenefitToken extends BaseToken with entitlement management features.
type BenefitToken struct {
	*BaseToken
	ledger  StateRW
	records map[Hash]BenefitRecord
	mu      sync.RWMutex
}

// Base returns the embedded BaseToken instance.
func (bt *BenefitToken) Base() *BaseToken { return bt.BaseToken }

// NewBenefitToken creates an empty SYN3900 benefit token.
func NewBenefitToken(meta Metadata) *BenefitToken {
	return &BenefitToken{
		BaseToken: &BaseToken{id: deriveID(meta.Standard), meta: meta, balances: NewBalanceTable()},
		records:   make(map[Hash]BenefitRecord),
	}
}

// Issue allocates a benefit to the recipient with optional validity bounds.
func (bt *BenefitToken) Issue(recipient Address, amount uint64, validFrom, validUntil time.Time, conditions string) (Hash, error) {
	if bt.ledger == nil {
		return Hash{}, errors.New("ledger not initialised")
	}
	id := sha256.Sum256([]byte(recipient.String() + time.Now().String()))
	rec := BenefitRecord{
		ID:         id,
		Recipient:  recipient,
		Amount:     amount,
		ValidFrom:  validFrom.Unix(),
		ValidUntil: validUntil.Unix(),
		IssuedAt:   time.Now().Unix(),
		Conditions: conditions,
	}
	bt.mu.Lock()
	bt.records[id] = rec
	bt.mu.Unlock()
	if err := bt.Mint(recipient, amount); err != nil {
		return Hash{}, err
	}
	raw, _ := json.Marshal(rec)
	_ = bt.ledger.SetState(bt.key(id), raw)
	return id, nil
}

// Claim marks a benefit as claimed if conditions are met.
func (bt *BenefitToken) Claim(id Hash, claimer Address) error {
	bt.mu.Lock()
	rec, ok := bt.records[id]
	if !ok {
		bt.mu.Unlock()
		return errors.New("unknown benefit")
	}
	if rec.Claimed || claimer != rec.Recipient {
		bt.mu.Unlock()
		return errors.New("invalid claim")
	}
	now := time.Now().Unix()
	if now < rec.ValidFrom || (rec.ValidUntil > 0 && now > rec.ValidUntil) {
		bt.mu.Unlock()
		return errors.New("benefit not valid")
	}
	rec.Claimed = true
	bt.records[id] = rec
	bt.mu.Unlock()
	raw, _ := json.Marshal(rec)
	_ = bt.ledger.SetState(bt.key(id), raw)
	return nil
}

// Record fetches a stored benefit record from memory.
func (bt *BenefitToken) Record(id Hash) (BenefitRecord, bool) {
	bt.mu.RLock()
	rec, ok := bt.records[id]
	bt.mu.RUnlock()
	return rec, ok
}

func (bt *BenefitToken) key(id Hash) []byte {
	return append([]byte("benefit:"), id[:]...)
}
