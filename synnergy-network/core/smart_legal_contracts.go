package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// SmartLegalRegistry manages Ricardian contracts and signer approvals.
type SmartLegalRegistry struct {
	ledger    *Ledger
	mu        sync.RWMutex
	contracts map[Address]*RicardianContract
	signers   map[Address]map[Address]time.Time
}

var (
	smartLegalOnce sync.Once
	smartLegalReg  *SmartLegalRegistry
)

// InitSmartLegalContracts initialises the global registry using the provided ledger.
func InitSmartLegalContracts(led *Ledger) {
	smartLegalOnce.Do(func() {
		smartLegalReg = &SmartLegalRegistry{
			ledger:    led,
			contracts: make(map[Address]*RicardianContract),
			signers:   make(map[Address]map[Address]time.Time),
		}
	})
}

// RegisterAgreement stores a Ricardian contract in the registry and ledger.
func RegisterAgreement(rc RicardianContract) error {
	if smartLegalReg == nil {
		return errors.New("smart legal registry not initialised")
	}
	smartLegalReg.mu.Lock()
	defer smartLegalReg.mu.Unlock()

	if _, exists := smartLegalReg.contracts[rc.Address]; exists {
		return fmt.Errorf("agreement already exists")
	}
	if rc.Created.IsZero() {
		rc.Created = time.Now().UTC()
	}
	smartLegalReg.contracts[rc.Address] = &rc
	if smartLegalReg.ledger != nil {
		raw, err := json.Marshal(rc)
		if err != nil {
			return err
		}
		if err := smartLegalReg.ledger.SetState(legalAgreementKey(rc.Address), raw); err != nil {
			return err
		}
	}
	return nil
}

// SignAgreement records that a party has accepted the terms of a contract.
func SignAgreement(contract, party Address) error {
	if smartLegalReg == nil {
		return errors.New("smart legal registry not initialised")
	}
	if smartLegalReg.ledger != nil && !smartLegalReg.ledger.IsIDTokenHolder(party) {
		return fmt.Errorf("party %s not authorised", party)
	}
	smartLegalReg.mu.Lock()
	defer smartLegalReg.mu.Unlock()

	if smartLegalReg.signers[contract] == nil {
		smartLegalReg.signers[contract] = make(map[Address]time.Time)
	}
	smartLegalReg.signers[contract][party] = time.Now().UTC()
	return nil
}

// RevokeAgreement removes a previously recorded signature.
func RevokeAgreement(contract, party Address) error {
	if smartLegalReg == nil {
		return errors.New("smart legal registry not initialised")
	}
	smartLegalReg.mu.Lock()
	defer smartLegalReg.mu.Unlock()

	if m, ok := smartLegalReg.signers[contract]; ok {
		delete(m, party)
	}
	return nil
}

// AgreementInfo returns the Ricardian contract and list of signers.
func AgreementInfo(addr Address) (*RicardianContract, []Address, error) {
	if smartLegalReg == nil {
		return nil, nil, errors.New("smart legal registry not initialised")
	}
	smartLegalReg.mu.RLock()
	defer smartLegalReg.mu.RUnlock()

	rc, ok := smartLegalReg.contracts[addr]
	if !ok {
		return nil, nil, fmt.Errorf("agreement %s not found", addr)
	}
	var parties []Address
	if m, ok := smartLegalReg.signers[addr]; ok {
		for p := range m {
			parties = append(parties, p)
		}
	}
	return rc, parties, nil
}

// ListAgreements returns all registered agreements.
func ListAgreements() map[Address]*RicardianContract {
	if smartLegalReg == nil {
		return nil
	}
	smartLegalReg.mu.RLock()
	defer smartLegalReg.mu.RUnlock()

	out := make(map[Address]*RicardianContract, len(smartLegalReg.contracts))
	for a, rc := range smartLegalReg.contracts {
		out[a] = rc
	}
	return out
}

func legalAgreementKey(a Address) []byte { return append([]byte("legal:"), a.Bytes()...) }
