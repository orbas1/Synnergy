package core

import (
	"encoding/json"
	"errors"
	"sync"
)

// Regulator represents an approved regulatory authority
// allowed to enforce compliance policies on chain.
type Regulator struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Jurisdiction string `json:"jurisdiction"`
}

var (
	regulationMu sync.RWMutex
	regulators   map[string]Regulator
	regLedger    *Ledger
)

// InitRegulatory sets the ledger used for persistence. It must be
// called before any other regulatory management function.
func InitRegulatory(led *Ledger) {
	regulationMu.Lock()
	defer regulationMu.Unlock()
	regLedger = led
	if regulators == nil {
		regulators = make(map[string]Regulator)
	}
}

// RegisterRegulator records a new regulator in memory and on the ledger.
func RegisterRegulator(id, name, jurisdiction string) error {
	if id == "" || name == "" {
		return errors.New("id and name required")
	}
	regulationMu.Lock()
	defer regulationMu.Unlock()
	if _, ok := regulators[id]; ok {
		return errors.New("regulator exists")
	}
	r := Regulator{ID: id, Name: name, Jurisdiction: jurisdiction}
	regulators[id] = r
	if regLedger != nil {
		b, _ := json.Marshal(r)
		regLedger.SetState(regMgmtKey(id), b)
	}
	return nil
}

// GetRegulator retrieves a regulator from memory or ledger.
func GetRegulator(id string) (Regulator, bool) {
	regulationMu.RLock()
	r, ok := regulators[id]
	regulationMu.RUnlock()
	if ok {
		return r, true
	}
	if regLedger == nil {
		return Regulator{}, false
	}
	b, _ := regLedger.GetState(regMgmtKey(id))
	if len(b) == 0 {
		return Regulator{}, false
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return Regulator{}, false
	}
	regulationMu.Lock()
	regulators[id] = r
	regulationMu.Unlock()
	return r, true
}

// ListRegulators returns all currently registered regulators.
func ListRegulators() []Regulator {
	regulationMu.RLock()
	list := make([]Regulator, 0, len(regulators))
	for _, r := range regulators {
		list = append(list, r)
	}
	regulationMu.RUnlock()
	return list
}

// EvaluateRuleSet performs a minimal compliance check on a transaction.
// In this prototype it ensures every output recipient holds an ID token.
func EvaluateRuleSet(tx *Transaction) error {
	if tx == nil {
		return errors.New("nil tx")
	}
	regulationMu.RLock()
	led := regLedger
	regulationMu.RUnlock()
	if led == nil {
		return nil
	}
	for _, out := range tx.Outputs {
		if !led.IsIDTokenHolder(out.Address) {
			return errors.New("destination not verified")
		}
	}
	return nil
}

func regMgmtKey(id string) []byte { return []byte("reg:" + id) }
