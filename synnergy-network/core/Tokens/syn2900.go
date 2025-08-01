package Tokens

import "time"

// InsurancePolicy mirrors the structure in the core package but avoids
// direct dependencies to prevent cycles.
type InsurancePolicy struct {
	ID            string
	Holder        [20]byte
	Coverage      string
	Premium       uint64
	Payout        uint64
	Deductible    uint64
	CoverageLimit uint64
	StartDate     time.Time
	EndDate       time.Time
	Active        bool
}

// InsuranceToken exposes the SYN2900 insurance-specific methods.
type InsuranceToken interface {
	TokenInterfaces
	IssuePolicy(holder [20]byte, coverage string, premium, payout, deductible, limit uint64, start, end time.Time) (string, error)
	GetPolicy(id string) (InsurancePolicy, bool)
	UpdatePolicy(pol InsurancePolicy) error
	CancelPolicy(id string) error
	ClaimPolicy(id string) error
}
