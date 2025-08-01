package core

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"sync"
)

// ContractManager provides administrative lifecycle operations for
// deployed smart contracts. It persists metadata via the Ledger
// using well known key prefixes so state survives restarts.
//
// Functions are concurrency safe and integrate with the existing
// ContractRegistry and VM.

type ContractManager struct {
	ledger *Ledger
	reg    *ContractRegistry
	mu     sync.RWMutex
}

// Prefixes used when storing state in the ledger key/value store.
const (
	ownerPrefix  = "contract:owner:"
	pausedPrefix = "contract:paused:"
)

// NewContractManager wires the manager with the given ledger and
// contract registry.
func NewContractManager(led *Ledger, reg *ContractRegistry) *ContractManager {
	return &ContractManager{ledger: led, reg: reg}
}

// TransferOwnership assigns a new owner for the contract. The address
// is persisted in the ledger so it can be queried later.
func (cm *ContractManager) TransferOwnership(addr, newOwner Address) error {
	if cm.ledger == nil || cm.reg == nil {
		return errors.New("contract manager not initialised")
	}
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, ok := cm.reg.byAddr[addr]; !ok {
		return errors.New("contract not found")
	}
	return cm.ledger.SetState(ownerKey(addr), newOwner.Bytes())
}

// OwnerOf fetches the currently assigned owner of a contract. If no
// owner has been recorded an empty Address is returned.
func (cm *ContractManager) OwnerOf(addr Address) (Address, error) {
	if cm.ledger == nil {
		return Address{}, errors.New("ledger not available")
	}
	b, err := cm.ledger.GetState(ownerKey(addr))
	if err != nil {
		return Address{}, err
	}
	var out Address
	copy(out[:], b)
	return out, nil
}

// PauseContract marks the contract as paused. The VM should reject
// calls when this flag is set. The manager only records state.
func (cm *ContractManager) PauseContract(addr Address) error {
	if cm.ledger == nil {
		return errors.New("ledger not available")
	}
	return cm.ledger.SetState(pausedKey(addr), []byte{1})
}

// ResumeContract clears the paused flag.
func (cm *ContractManager) ResumeContract(addr Address) error {
	if cm.ledger == nil {
		return errors.New("ledger not available")
	}
	return cm.ledger.SetState(pausedKey(addr), []byte{0})
}

// IsPaused reports whether a contract is currently paused.
func (cm *ContractManager) IsPaused(addr Address) bool {
	if cm.ledger == nil {
		return false
	}
	b, err := cm.ledger.GetState(pausedKey(addr))
	return err == nil && len(b) > 0 && b[0] == 1
}

// UpgradeContract replaces the bytecode for a deployed contract and
// updates the registry entry. Existing paused state is preserved.
func (cm *ContractManager) UpgradeContract(addr Address, code []byte, gas uint64) error {
	if cm.ledger == nil || cm.reg == nil {
		return errors.New("contract manager not initialised")
	}
	cm.mu.Lock()
	defer cm.mu.Unlock()
	sc, ok := cm.reg.byAddr[addr]
	if !ok {
		return errors.New("contract not found")
	}
	if cm.IsPaused(addr) {
		return errors.New("contract is paused")
	}
	hash := sha256.Sum256(code)
	sc.Bytecode = code
	sc.CodeHash = hash
	sc.GasLimit = gas
	if err := cm.ledger.SetState(contractKey(addr), code); err != nil {
		return err
	}
	return nil
}

// ContractInfo returns a JSON blob describing the contract including
// owner and paused status.
func (cm *ContractManager) ContractInfo(addr Address) ([]byte, error) {
	if cm.reg == nil {
		return nil, errors.New("registry not initialised")
	}
	cm.mu.RLock()
	sc, ok := cm.reg.byAddr[addr]
	cm.mu.RUnlock()
	if !ok {
		return nil, errors.New("contract not found")
	}
	owner, _ := cm.OwnerOf(addr)
	info := struct {
		*SmartContract
		Owner  Address `json:"owner"`
		Paused bool    `json:"paused"`
	}{sc, owner, cm.IsPaused(addr)}
	return json.MarshalIndent(info, "", "  ")
}

func ownerKey(addr Address) []byte  { return append([]byte(ownerPrefix), addr.Bytes()...) }
func pausedKey(addr Address) []byte { return append([]byte(pausedPrefix), addr.Bytes()...) }
