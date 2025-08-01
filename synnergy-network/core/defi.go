package core

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// DeFiManager exposes basic decentralised finance helpers. The
// implementation is intentionally lightweight and stores all
// state in the main ledger so consensus and the VM can share data.

type DeFiManager struct {
	mu     sync.Mutex
	ledger *Ledger
}

func NewDeFiManager(led *Ledger) *DeFiManager { return &DeFiManager{ledger: led} }

// helper utilities
func (dm *DeFiManager) store(key []byte, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return dm.ledger.SetState(key, b)
}

func (dm *DeFiManager) load(key []byte, out any) error {
	b, err := dm.ledger.GetState(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func (dm *DeFiManager) mint(to Address, token string, amt uint64) error {
	return dm.ledger.MintToken(to, token, amt)
}

// ------------------------------------------------------------------
// Insurance
// ------------------------------------------------------------------

type InsurancePolicy struct {
	ID      Hash    `json:"id"`
	Holder  Address `json:"holder"`
	Premium uint64  `json:"premium"`
	Payout  uint64  `json:"payout"`
	Active  bool    `json:"active"`
	Created int64   `json:"created"`
}

func (dm *DeFiManager) CreateInsurance(id Hash, holder Address, premium, payout uint64) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	k := append([]byte("ins:"), id[:]...)
	if ok, _ := dm.ledger.HasState(k); ok {
		return fmt.Errorf("exists")
	}
	pol := InsurancePolicy{ID: id, Holder: holder, Premium: premium, Payout: payout, Active: true, Created: time.Now().Unix()}
	if err := dm.ledger.Transfer(holder, BurnAddress, premium); err != nil {
		return err
	}
	return dm.store(k, pol)
}

func (dm *DeFiManager) ClaimInsurance(id Hash) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	k := append([]byte("ins:"), id[:]...)
	var pol InsurancePolicy
	if err := dm.load(k, &pol); err != nil {
		return err
	}
	if !pol.Active {
		return fmt.Errorf("not active")
	}
	pol.Active = false
	if err := dm.mint(pol.Holder, "SYNTHRON", pol.Payout); err != nil {
		return err
	}
	return dm.store(k, pol)
}
