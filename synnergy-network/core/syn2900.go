package core

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TokenInsurancePolicy represents a blockchain based insurance policy.
type TokenInsurancePolicy struct {
	ID            string    `json:"id"`
	Holder        Address   `json:"holder"`
	Coverage      string    `json:"coverage"`
	Premium       uint64    `json:"premium"`
	Payout        uint64    `json:"payout"`
	Deductible    uint64    `json:"deductible"`
	CoverageLimit uint64    `json:"coverage_limit"`
	StartDate     time.Time `json:"start"`
	EndDate       time.Time `json:"end"`
	Active        bool      `json:"active"`
}

// InsuranceToken extends BaseToken with policy management.
type InsuranceToken struct {
	*BaseToken
	mu       sync.RWMutex
	policies map[string]TokenInsurancePolicy
}

// NewInsuranceToken constructs an empty insurance token.
func NewInsuranceToken(meta Metadata) *InsuranceToken {
	bt := &BaseToken{id: deriveID(meta.Standard), meta: meta, balances: NewBalanceTable()}
	return &InsuranceToken{BaseToken: bt, policies: make(map[string]TokenInsurancePolicy)}
}

func (it *InsuranceToken) policyKey(id string) []byte {
	return []byte("pol:" + id)
}

// IssuePolicy creates a new insurance policy and mints one token to the holder.
func (it *InsuranceToken) IssuePolicy(holder Address, coverage string, premium, payout, deductible, limit uint64, start, end time.Time) (string, error) {
	it.mu.Lock()
	defer it.mu.Unlock()
	if it.ledger == nil {
		return "", fmt.Errorf("ledger not initialised")
	}
	pid := uuid.New().String()
	pol := TokenInsurancePolicy{ID: pid, Holder: holder, Coverage: coverage, Premium: premium, Payout: payout, Deductible: deductible, CoverageLimit: limit, StartDate: start, EndDate: end, Active: true}
	blob, _ := json.Marshal(pol)
	if err := it.ledger.SetState(it.policyKey(pid), blob); err != nil {
		return "", err
	}
	it.policies[pid] = pol
	if err := it.Mint(holder, 1); err != nil {
		return "", err
	}
	return pid, nil
}

// GetPolicy loads the policy by ID.
func (it *InsuranceToken) GetPolicy(id string) (TokenInsurancePolicy, bool) {
	it.mu.RLock()
	defer it.mu.RUnlock()
	pol, ok := it.policies[id]
	if ok {
		return pol, true
	}
	if it.ledger != nil {
		if blob, err := it.ledger.GetState(it.policyKey(id)); err == nil {
			_ = json.Unmarshal(blob, &pol)
			return pol, true
		}
	}
	return TokenInsurancePolicy{}, false
}

// UpdatePolicy writes updated policy data to the ledger.
func (it *InsuranceToken) UpdatePolicy(pol TokenInsurancePolicy) error {
	it.mu.Lock()
	defer it.mu.Unlock()
	if it.ledger == nil {
		return fmt.Errorf("ledger not initialised")
	}
	blob, _ := json.Marshal(pol)
	if err := it.ledger.SetState(it.policyKey(pol.ID), blob); err != nil {
		return err
	}
	it.policies[pol.ID] = pol
	return nil
}

// CancelPolicy marks the policy inactive.
func (it *InsuranceToken) CancelPolicy(id string) error {
	it.mu.Lock()
	defer it.mu.Unlock()
	pol, ok := it.policies[id]
	if !ok {
		if blob, err := it.ledger.GetState(it.policyKey(id)); err == nil {
			_ = json.Unmarshal(blob, &pol)
		} else {
			return fmt.Errorf("policy not found")
		}
	}
	pol.Active = false
	blob, _ := json.Marshal(pol)
	if err := it.ledger.SetState(it.policyKey(id), blob); err != nil {
		return err
	}
	it.policies[id] = pol
	return nil
}

// ClaimPolicy pays out the insurance and marks policy inactive.
func (it *InsuranceToken) ClaimPolicy(id string) error {
	it.mu.Lock()
	defer it.mu.Unlock()
	if it.ledger == nil {
		return fmt.Errorf("ledger not initialised")
	}
	pol, ok := it.policies[id]
	if !ok {
		if blob, err := it.ledger.GetState(it.policyKey(id)); err == nil {
			_ = json.Unmarshal(blob, &pol)
		} else {
			return fmt.Errorf("policy not found")
		}
	}
	if !pol.Active {
		return fmt.Errorf("policy inactive")
	}
	pol.Active = false
	blob, _ := json.Marshal(pol)
	if err := it.ledger.SetState(it.policyKey(id), blob); err != nil {
		return err
	}
	it.policies[id] = pol
	if err := it.Mint(pol.Holder, pol.Payout); err != nil {
		return err
	}
	return nil
}
