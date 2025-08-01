//go:build tokens
// +build tokens

package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// VerificationRecord captures third-party verification details for a project.
// It is defined locally to avoid cyclic imports between core and token packages.
type VerificationRecord struct {
	ID        string    `json:"id"`
	Verifier  Address   `json:"verifier"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

// CarbonProject represents a verified carbon offset project recorded on chain.
type CarbonProject struct {
	ID            uint64               `json:"id"`
	Name          string               `json:"name"`
	Owner         Address              `json:"owner"`
	Total         uint64               `json:"total"`
	Remaining     uint64               `json:"remaining"`
	Verified      bool                 `json:"verified"`
	Verifications []VerificationRecord `json:"verifications"`
}

// CarbonEngine manages carbon credit issuance and retirement. Projects are
// stored on the ledger under the "ccs:proj:" prefix. Credits are minted using
// the built‑in SYN-CO2 token.
type CarbonEngine struct {
	ledger  StateRW
	mu      sync.RWMutex
	nextID  uint64
	tokenID TokenID
}

var (
	carbon     *CarbonEngine
	carbonOnce sync.Once
)

// CarbonCreditTokenID is the TokenID used for carbon credit issuance.
const CarbonCreditTokenID TokenID = TokenID(0x53000000 | uint32(StdSYN200)<<8)

// InitCarbonEngine initialises the singleton engine.
func InitCarbonEngine(led StateRW) {
	carbonOnce.Do(func() {
		carbon = &CarbonEngine{ledger: led, tokenID: CarbonCreditTokenID}
		if b, err := led.GetState([]byte("ccs:nextID")); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &carbon.nextID)
		}
	})
}

// Carbon returns the global engine instance.
func Carbon() *CarbonEngine { return carbon }

func (e *CarbonEngine) projectKey(id uint64) []byte {
	return []byte(fmt.Sprintf("ccs:proj:%d", id))
}

// RegisterProject creates a new carbon offset project owned by `owner`.
func (e *CarbonEngine) RegisterProject(owner Address, name string, total uint64) (uint64, error) {
	if name == "" || total == 0 {
		return 0, errors.New("invalid project parameters")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nextID++
	id := e.nextID
	proj := CarbonProject{ID: id, Name: name, Owner: owner, Total: total, Remaining: total}
	blob, _ := json.Marshal(proj)
	if err := e.ledger.SetState(e.projectKey(id), blob); err != nil {
		e.nextID--
		return 0, err
	}
	b, _ := json.Marshal(e.nextID)
	_ = e.ledger.SetState([]byte("ccs:nextID"), b)
	return id, nil
}

func (e *CarbonEngine) getProject(id uint64) (*CarbonProject, error) {
	b, err := e.ledger.GetState(e.projectKey(id))
	if err != nil || len(b) == 0 {
		return nil, fmt.Errorf("project %d not found", id)
	}
	var p CarbonProject
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// IssueCredits mints SYN‑CO2 tokens to `to` from the given project.
func (e *CarbonEngine) IssueCredits(id uint64, to Address, amount uint64) error {
	if amount == 0 {
		return errors.New("amount > 0")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	proj, err := e.getProject(id)
	if err != nil {
		return err
	}
	if proj.Remaining < amount {
		return fmt.Errorf("insufficient project credits")
	}
	proj.Remaining -= amount
	blob, _ := json.Marshal(proj)
	if err := e.ledger.SetState(e.projectKey(id), blob); err != nil {
		return err
	}
	tok, ok := GetToken(e.tokenID)
	if !ok {
		return fmt.Errorf("carbon credit token not found")
	}
	return tok.Mint(to, amount)
}

// RetireCredits burns SYN‑CO2 tokens from the holder's balance.
func (e *CarbonEngine) RetireCredits(holder Address, amount uint64) error {
	if amount == 0 {
		return errors.New("amount > 0")
	}
	tok, ok := GetToken(e.tokenID)
	if !ok {
		return fmt.Errorf("carbon credit token not found")
	}
	return tok.Burn(holder, amount)
}

// AddVerification records a verification entry for the specified project.
func (e *CarbonEngine) AddVerification(id uint64, ver VerificationRecord) error {
	if ver.ID == "" {
		return errors.New("verification id required")
	}
	ver.Timestamp = time.Now().UTC()
	e.mu.Lock()
	defer e.mu.Unlock()
	proj, err := e.getProject(id)
	if err != nil {
		return err
	}
	for _, existing := range proj.Verifications {
		if existing.ID == ver.ID {
			return fmt.Errorf("verification %s exists", ver.ID)
		}
	}
	proj.Verifications = append(proj.Verifications, ver)
	if ver.Status == "verified" {
		proj.Verified = true
	}
	blob, _ := json.Marshal(proj)
	return e.ledger.SetState(e.projectKey(id), blob)
}

// ListVerifications returns verification records for a project.
func (e *CarbonEngine) ListVerifications(id uint64) ([]VerificationRecord, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	proj, err := e.getProject(id)
	if err != nil {
		return nil, err
	}
	out := make([]VerificationRecord, len(proj.Verifications))
	copy(out, proj.Verifications)
	return out, nil
}

// ProjectInfo returns a project by id.
func (e *CarbonEngine) ProjectInfo(id uint64) (*CarbonProject, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	proj, err := e.getProject(id)
	if err != nil {
		return nil, false
	}
	return proj, true
}

// ListProjects enumerates all projects in the ledger.
func (e *CarbonEngine) ListProjects() ([]CarbonProject, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	iter := e.ledger.PrefixIterator([]byte("ccs:proj:"))
	var list []CarbonProject
	for iter.Next() {
		var p CarbonProject
		if err := json.Unmarshal(iter.Value(), &p); err != nil {
			continue
		}
		list = append(list, p)
	}
	return list, iter.Error()
}

// End of carbon_credit_system.go
