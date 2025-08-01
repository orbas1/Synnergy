package core

import (
	"encoding/json"
	"fmt"
	"time"
)

// GovernanceContract holds metadata about a governance smart contract.
type GovernanceContract struct {
	Address Address   `json:"address"`
	Name    string    `json:"name"`
	Enabled bool      `json:"enabled"`
	Updated time.Time `json:"updated"`
}

// GovernanceManager coordinates governance contracts and proposals.
type GovernanceManager struct {
	ledger *Ledger
}

// NewGovernanceManager initialises a manager bound to a ledger.
// It also wires the global governance ledger reference so helper
// functions in governance.go operate on the same instance.
func NewGovernanceManager(led *Ledger) *GovernanceManager {
	if led != nil {
		ledger = led
	}
	return &GovernanceManager{ledger: led}
}

// RegisterGovContract stores a governance contract definition.
func (gm *GovernanceManager) RegisterGovContract(addr Address, name string) error {
	if gm.ledger == nil {
		return fmt.Errorf("governance manager: nil ledger")
	}
	c := GovernanceContract{Address: addr, Name: name, Enabled: true, Updated: time.Now().UTC()}
	raw, err := json.Marshal(c)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("gov:contract:%x", addr[:])
	return CurrentStore().Set([]byte(key), raw)
}

// GetGovContract retrieves a governance contract by address.
func (gm *GovernanceManager) GetGovContract(addr Address) (GovernanceContract, error) {
	key := fmt.Sprintf("gov:contract:%x", addr[:])
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil {
		return GovernanceContract{}, err
	}
	var c GovernanceContract
	if err := json.Unmarshal(raw, &c); err != nil {
		return GovernanceContract{}, err
	}
	return c, nil
}

// ListGovContracts returns all registered governance contracts.
func (gm *GovernanceManager) ListGovContracts() ([]GovernanceContract, error) {
	it := CurrentStore().Iterator([]byte("gov:contract:"), nil)
	var out []GovernanceContract
	for it.Next() {
		var c GovernanceContract
		if err := json.Unmarshal(it.Value(), &c); err == nil {
			out = append(out, c)
		}
	}
	return out, it.Error()
}

// EnableGovContract toggles whether a contract is active.
func (gm *GovernanceManager) EnableGovContract(addr Address, enable bool) error {
	c, err := gm.GetGovContract(addr)
	if err != nil {
		return err
	}
	c.Enabled = enable
	c.Updated = time.Now().UTC()
	raw, _ := json.Marshal(c)
	key := fmt.Sprintf("gov:contract:%x", addr[:])
	return CurrentStore().Set([]byte(key), raw)
}

// DeleteGovContract removes a contract from state.
func (gm *GovernanceManager) DeleteGovContract(addr Address) error {
	key := fmt.Sprintf("gov:contract:%x", addr[:])
	return CurrentStore().Delete([]byte(key))
}
