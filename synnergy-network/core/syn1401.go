package core

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// InvestmentRecord stores state for a SYN1401 investment token issuance.
type InvestmentRecord struct {
	ID           string    `json:"id"`
	Owner        Address   `json:"owner"`
	Principal    uint64    `json:"principal"`
	InterestRate float64   `json:"interest_rate"`
	StartDate    time.Time `json:"start_date"`
	MaturityDate time.Time `json:"maturity_date"`
	Accrued      uint64    `json:"accrued"`
	Redeemed     bool      `json:"redeemed"`
}

// InvestmentManager manages SYN1401 investment tokens through the ledger.
type InvestmentManager struct {
	Ledger StateRW
	mu     sync.Mutex
}

// NewInvestmentManager returns a manager bound to the given ledger.
func NewInvestmentManager(led StateRW) *InvestmentManager { return &InvestmentManager{Ledger: led} }

func invKey(id string) []byte { return []byte("syn1401:" + id) }

// Issue registers a new investment record.
func (m *InvestmentManager) Issue(id string, owner Address, principal uint64, rate float64, maturity time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ok, _ := m.Ledger.HasState(invKey(id)); ok {
		return fmt.Errorf("investment %s exists", id)
	}
	rec := InvestmentRecord{ID: id, Owner: owner, Principal: principal, InterestRate: rate, StartDate: time.Now().UTC(), MaturityDate: maturity}
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return m.Ledger.SetState(invKey(id), data)
}

// Accrue updates accrued interest up to the current day.
func (m *InvestmentManager) Accrue(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	raw, err := m.Ledger.GetState(invKey(id))
	if err != nil || raw == nil {
		return fmt.Errorf("investment not found")
	}
	var rec InvestmentRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return err
	}
	if rec.Redeemed {
		return fmt.Errorf("already redeemed")
	}
	days := int(time.Since(rec.StartDate).Hours() / 24)
	accrued := uint64(float64(rec.Principal) * rec.InterestRate * float64(days) / 365)
	if accrued <= rec.Accrued {
		return nil
	}
	rec.Accrued = accrued
	upd, err := json.Marshal(&rec)
	if err != nil {
		return err
	}
	return m.Ledger.SetState(invKey(id), upd)
}

// Redeem pays out principal and accrued interest at maturity.
func (m *InvestmentManager) Redeem(id string, to Address) (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	raw, err := m.Ledger.GetState(invKey(id))
	if err != nil || raw == nil {
		return 0, fmt.Errorf("investment not found")
	}
	var rec InvestmentRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return 0, err
	}
	if rec.Redeemed {
		return 0, fmt.Errorf("already redeemed")
	}
	if time.Now().Before(rec.MaturityDate) {
		return 0, fmt.Errorf("not matured")
	}
	total := rec.Principal + rec.Accrued
	if err := m.Ledger.Transfer(ModuleAddress("syn1401"), to, total); err != nil {
		return 0, err
	}
	rec.Redeemed = true
	upd, err := json.Marshal(&rec)
	if err != nil {
		return 0, err
	}
	if err := m.Ledger.SetState(invKey(id), upd); err != nil {
		return 0, err
	}
	return total, nil
}

// Get returns investment details by id.
func (m *InvestmentManager) Get(id string) (InvestmentRecord, bool, error) {
	raw, err := m.Ledger.GetState(invKey(id))
	if err != nil || raw == nil {
		return InvestmentRecord{}, false, err
	}
	var rec InvestmentRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return InvestmentRecord{}, false, err
	}
	return rec, true, nil
}

// List returns all investment records.
func (m *InvestmentManager) List() ([]InvestmentRecord, error) {
	it := m.Ledger.PrefixIterator([]byte("syn1401:"))
	var out []InvestmentRecord
	for it.Next() {
		var rec InvestmentRecord
		if err := json.Unmarshal(it.Value(), &rec); err == nil {
			out = append(out, rec)
		}
	}
	if err := it.Error(); err != nil {
		return nil, err
	}
	return out, nil
}

// registerSYN1401Opcodes is wired via opcode_dispatcher.go catalogue.
func registerSYN1401Opcodes() {}
