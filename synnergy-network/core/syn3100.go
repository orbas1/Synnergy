package core

import (
	"errors"
	"sync"
	"time"
)

// EmploymentContractMeta stores employment token contract data
// associated with SYN3100 tokens.
type EmploymentContractMeta struct {
	ContractID string
	Employer   Address
	Employee   Address
	Position   string
	Salary     uint64
	Benefits   string
	Start      time.Time
	End        time.Time
	Active     bool
}

// EmploymentToken implements the SYN3100 employment token standard.
type EmploymentToken struct {
	BaseToken
	contracts map[string]EmploymentContractMeta
	mu        sync.RWMutex
}

// CreateContract registers a new employment contract metadata entry.
func (et *EmploymentToken) CreateContract(meta EmploymentContractMeta) error {
	et.mu.Lock()
	defer et.mu.Unlock()
	if et.contracts == nil {
		et.contracts = make(map[string]EmploymentContractMeta)
	}
	if _, exists := et.contracts[meta.ContractID]; exists {
		return errors.New("contract exists")
	}
	meta.Active = true
	et.contracts[meta.ContractID] = meta
	return nil
}

// PaySalary transfers the salary amount from employer to employee.
func (et *EmploymentToken) PaySalary(contractID string) error {
	et.mu.RLock()
	meta, ok := et.contracts[contractID]
	et.mu.RUnlock()
	if !ok {
		return errors.New("unknown contract")
	}
	if !meta.Active {
		return errors.New("inactive contract")
	}
	if err := et.Transfer(meta.Employer, meta.Employee, meta.Salary); err != nil {
		return err
	}
	return nil
}

// UpdateBenefits modifies the benefits string for a contract.
func (et *EmploymentToken) UpdateBenefits(contractID, benefits string) error {
	et.mu.Lock()
	defer et.mu.Unlock()
	meta, ok := et.contracts[contractID]
	if !ok {
		return errors.New("unknown contract")
	}
	meta.Benefits = benefits
	et.contracts[contractID] = meta
	return nil
}

// TerminateContract marks the contract as inactive.
func (et *EmploymentToken) TerminateContract(contractID string) error {
	et.mu.Lock()
	defer et.mu.Unlock()
	meta, ok := et.contracts[contractID]
	if !ok {
		return errors.New("unknown contract")
	}
	meta.Active = false
	et.contracts[contractID] = meta
	return nil
}

// GetContract retrieves metadata for a contract ID.
func (et *EmploymentToken) GetContract(contractID string) (EmploymentContractMeta, bool) {
	et.mu.RLock()
	defer et.mu.RUnlock()
	meta, ok := et.contracts[contractID]
	return meta, ok
}
