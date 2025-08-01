package core

import (
	"fmt"
	"sync"
	"time"

	Tokens "synnergy-network/core/Tokens"
)

// EducationToken implements the SYN1900 education credit standard.
type EducationToken struct {
	*BaseToken
	records map[string]Tokens.EducationCreditMetadata
	mu      sync.RWMutex
}

// NewEducationToken creates a new education credit token.
func NewEducationToken(meta Metadata, ledger *Ledger, gas GasCalculator) *EducationToken {
	bt := &BaseToken{
		id:       deriveID(meta.Standard),
		meta:     meta,
		balances: NewBalanceTable(),
		ledger:   ledger,
		gas:      gas,
	}
	return &EducationToken{BaseToken: bt, records: make(map[string]Tokens.EducationCreditMetadata)}
}

// IssueCredit registers a new education credit.
func (et *EducationToken) IssueCredit(rec Tokens.EducationCreditMetadata) error {
	et.mu.Lock()
	defer et.mu.Unlock()
	if rec.CreditID == "" {
		return fmt.Errorf("credit id required")
	}
	rec.IssueDate = time.Now().UTC()
	et.records[rec.CreditID] = rec
	return nil
}

// VerifyCredit checks if a credit exists.
func (et *EducationToken) VerifyCredit(id string) bool {
	et.mu.RLock()
	defer et.mu.RUnlock()
	_, ok := et.records[id]
	return ok
}

// RevokeCredit removes a credit record.
func (et *EducationToken) RevokeCredit(id string) error {
	et.mu.Lock()
	defer et.mu.Unlock()
	if _, ok := et.records[id]; !ok {
		return fmt.Errorf("credit not found")
	}
	delete(et.records, id)
	return nil
}

// GetCredit retrieves a credit by ID.
func (et *EducationToken) GetCredit(id string) (Tokens.EducationCreditMetadata, bool) {
	et.mu.RLock()
	defer et.mu.RUnlock()
	rec, ok := et.records[id]
	return rec, ok
}

// ListCredits returns all credits for a recipient.
func (et *EducationToken) ListCredits(recipient string) []Tokens.EducationCreditMetadata {
	et.mu.RLock()
	defer et.mu.RUnlock()
	var out []Tokens.EducationCreditMetadata
	for _, r := range et.records {
		if r.Recipient == recipient {
			out = append(out, r)
		}
	}
	return out
}

var _ Tokens.EducationCreditTokenInterface = (*EducationToken)(nil)
