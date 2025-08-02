package Tokens

import "time"

// LifePolicy defines the metadata for a SYN2800 life insurance token.
type LifePolicy struct {
	PolicyID       string
	Insured        Address
	Beneficiary    Address
	Premium        uint64
	CoverageAmount uint64
	StartDate      time.Time
	EndDate        time.Time
	Active         bool
}

// SYN2800Token implements the SYN2800 life insurance token standard.
type SYN2800Token struct {
	Policy LifePolicy
}

// Meta returns the policy metadata.
func (t *SYN2800Token) Meta() any { return t.Policy }

// UpdatePremium sets the premium for the policy.
func (t *SYN2800Token) UpdatePremium(p uint64) { t.Policy.Premium = p }

// Activate marks the policy as active.
func (t *SYN2800Token) Activate() { t.Policy.Active = true }

// Deactivate marks the policy as inactive.
func (t *SYN2800Token) Deactivate() { t.Policy.Active = false }

// IsActive indicates whether the policy is currently active.
func (t *SYN2800Token) IsActive() bool { return t.Policy.Active }

// ProcessClaim reduces the coverage amount if the claim is valid and returns true.
func (t *SYN2800Token) ProcessClaim(amount uint64) bool {
	if !t.Policy.Active || amount > t.Policy.CoverageAmount {
		return false
	}
	t.Policy.CoverageAmount -= amount
	return true
}

// NewSYN2800Token constructs a new life insurance token instance.
func NewSYN2800Token(pol LifePolicy) *SYN2800Token {
	pol.StartDate = pol.StartDate.UTC()
	pol.EndDate = pol.EndDate.UTC()
	return &SYN2800Token{Policy: pol}
}
