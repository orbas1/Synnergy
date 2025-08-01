package core

import (
	"fmt"
	"sync"
	"time"
)

// ServiceTier defines access tiers for SYN500 utility tokens.
type ServiceTier uint8

const (
	TierBasic ServiceTier = iota
	TierPremium
	TierVIP
)

// AccessInfo tracks rights and rewards for a holder.
type AccessInfo struct {
	Tier         ServiceTier
	MaxUsage     uint64
	UsageCount   uint64
	Expiry       time.Time
	RewardPoints uint64
}

// SYN500Token represents the SYN500 utility token standard.
type SYN500Token struct {
	*BaseToken
	access map[Address]*AccessInfo
	mu     sync.RWMutex
}

// NewSYN500Token creates a new SYN500 token with metadata and initial balances.
func NewSYN500Token(meta Metadata, init map[Address]uint64) (*SYN500Token, error) {
	tok, err := (Factory{}).Create(meta, init)
	if err != nil {
		return nil, err
	}
	ut := &SYN500Token{
		BaseToken: tok.(*BaseToken),
		access:    make(map[Address]*AccessInfo),
	}
	return ut, nil
}

// GrantAccess sets access rights for an address.
func (ut *SYN500Token) GrantAccess(addr Address, tier ServiceTier, max uint64, expiry time.Time) {
	ut.mu.Lock()
	defer ut.mu.Unlock()
	ut.access[addr] = &AccessInfo{Tier: tier, MaxUsage: max, Expiry: expiry}
}

// UpdateAccess modifies existing access rights.
func (ut *SYN500Token) UpdateAccess(addr Address, tier ServiceTier, max uint64, expiry time.Time) {
	ut.mu.Lock()
	defer ut.mu.Unlock()
	if info, ok := ut.access[addr]; ok {
		info.Tier = tier
		info.MaxUsage = max
		info.Expiry = expiry
	} else {
		ut.access[addr] = &AccessInfo{Tier: tier, MaxUsage: max, Expiry: expiry}
	}
}

// RevokeAccess removes access rights for an address.
func (ut *SYN500Token) RevokeAccess(addr Address) {
	ut.mu.Lock()
	defer ut.mu.Unlock()
	delete(ut.access, addr)
}

// RecordUsage increments usage count and awards points.
func (ut *SYN500Token) RecordUsage(addr Address, points uint64) error {
	ut.mu.Lock()
	defer ut.mu.Unlock()
	info, ok := ut.access[addr]
	if !ok {
		return fmt.Errorf("no access for address")
	}
	if !info.Expiry.IsZero() && time.Now().After(info.Expiry) {
		return fmt.Errorf("access expired")
	}
	if info.MaxUsage > 0 && info.UsageCount >= info.MaxUsage {
		return fmt.Errorf("usage limit reached")
	}
	info.UsageCount++
	info.RewardPoints += points
	return nil
}

// RedeemReward deducts points from a holder.
func (ut *SYN500Token) RedeemReward(addr Address, points uint64) error {
	ut.mu.Lock()
	defer ut.mu.Unlock()
	info, ok := ut.access[addr]
	if !ok {
		return fmt.Errorf("no access for address")
	}
	if info.RewardPoints < points {
		return fmt.Errorf("insufficient points")
	}
	info.RewardPoints -= points
	return nil
}

// RewardBalance returns current reward points.
func (ut *SYN500Token) RewardBalance(addr Address) uint64 {
	ut.mu.RLock()
	defer ut.mu.RUnlock()
	if info, ok := ut.access[addr]; ok {
		return info.RewardPoints
	}
	return 0
}

// Usage returns the usage count for an address.
func (ut *SYN500Token) Usage(addr Address) uint64 {
	ut.mu.RLock()
	defer ut.mu.RUnlock()
	if info, ok := ut.access[addr]; ok {
		return info.UsageCount
	}
	return 0
}

// AccessInfoOf retrieves access details for an address.
func (ut *SYN500Token) AccessInfoOf(addr Address) (AccessInfo, bool) {
	ut.mu.RLock()
	defer ut.mu.RUnlock()
	info, ok := ut.access[addr]
	if !ok {
		return AccessInfo{}, false
	}
	return *info, true
}
