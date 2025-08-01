package core

import (
	"fmt"
	"sync"
	"time"
)

// MusicInfo captures the metadata for a music asset represented by SYN1600.
type MusicInfo struct {
	SongTitle   string
	Artist      string
	Album       string
	ReleaseDate time.Time
}

// RoyaltyShare defines the percentage of royalties owed to a recipient.
type RoyaltyShare struct {
	Recipient Address
	Percent   uint8 // 0-100
}

// RoyaltyEvent records a royalty distribution or revenue addition.
type RoyaltyEvent struct {
	Amount    uint64
	Timestamp time.Time
	TxID      string
}

// OwnershipChange tracks transfers of the music royalty token itself.
type OwnershipChange struct {
	From Address
	To   Address
	Time time.Time
}

// SYN1600Token implements the music royalty token standard.
type SYN1600Token struct {
	*BaseToken

	Info   MusicInfo
	Shares []RoyaltyShare

	revenue []RoyaltyEvent
	history []OwnershipChange
	mu      sync.RWMutex
}

// NewSYN1600Token constructs a music royalty token with initial holders.
func NewSYN1600Token(meta Metadata, init map[Address]uint64, info MusicInfo, shares []RoyaltyShare) (*SYN1600Token, error) {
	bt := &BaseToken{id: deriveID(meta.Standard), meta: meta, balances: NewBalanceTable()}
	for a, v := range init {
		bt.balances.Set(bt.id, a, v)
		bt.meta.TotalSupply += v
	}
	tok := &SYN1600Token{
		BaseToken: bt,
		Info:      info,
		Shares:    shares,
	}
	return tok, nil
}

// AddRevenue records incoming royalty revenue.
func (t *SYN1600Token) AddRevenue(amount uint64, txID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.revenue = append(t.revenue, RoyaltyEvent{Amount: amount, Timestamp: time.Now().UTC(), TxID: txID})
}

// RevenueHistory returns all recorded revenue events.
func (t *SYN1600Token) RevenueHistory() []RoyaltyEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]RoyaltyEvent, len(t.revenue))
	copy(out, t.revenue)
	return out
}

// DistributeRoyalties mints new tokens according to the configured royalty shares.
func (t *SYN1600Token) DistributeRoyalties(amount uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.Shares) == 0 {
		return fmt.Errorf("no royalty shares configured")
	}
	for _, s := range t.Shares {
		share := amount * uint64(s.Percent) / 100
		if err := t.Mint(s.Recipient, share); err != nil {
			return err
		}
	}
	t.revenue = append(t.revenue, RoyaltyEvent{Amount: amount, Timestamp: time.Now().UTC(), TxID: ""})
	return nil
}

// UpdateInfo updates the music asset metadata.
func (t *SYN1600Token) UpdateInfo(info MusicInfo) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Info = info
}

// RecordOwnership logs a change of ownership for auditing purposes.
func (t *SYN1600Token) RecordOwnership(from, to Address) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.history = append(t.history, OwnershipChange{From: from, To: to, Time: time.Now().UTC()})
}

// OwnershipHistory returns the historic transfers of this token.
func (t *SYN1600Token) OwnershipHistory() []OwnershipChange {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]OwnershipChange, len(t.history))
	copy(out, t.history)
	return out
}
