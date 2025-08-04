package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// EmploymentContract represents a simple on-chain employment agreement.
type EmploymentContract struct {
	ID            string    `json:"id"`
	Employer      Address   `json:"employer"`
	Employee      Address   `json:"employee"`
	SalaryPerHour uint64    `json:"salary_per_hour"`
	HoursWorked   uint32    `json:"hours_worked"`
	Start         time.Time `json:"start"`
	End           time.Time `json:"end"`
	EmployerSign  bool      `json:"employer_sign"`
	EmployeeSign  bool      `json:"employee_sign"`
	Paid          bool      `json:"paid"`
}

// EmploymentRegistry stores contracts in the ledger key/value store.
type EmploymentRegistry struct {
	led    *Ledger
	mu     sync.RWMutex
	nextID uint64
}

var (
	empOnce sync.Once
	empReg  *EmploymentRegistry
)

// InitEmployment initialises the global employment registry.
func InitEmployment(led *Ledger) {
	empOnce.Do(func() {
		empReg = &EmploymentRegistry{led: led, nextID: 1}
	})
}

// Employment returns the current registry instance.
func Employment() *EmploymentRegistry { return empReg }

// employmentToken attempts to locate the canonical SYN3100 token.
func employmentToken() *EmploymentToken {
	for _, t := range GetRegistryTokens() {
		if t.Meta().Standard == StdSYN3100 {
			if et, ok := any(t).(*EmploymentToken); ok {
				return et
			}
		}
	}
	return nil
}

// CreateJob creates a new employment contract.
func (r *EmploymentRegistry) CreateJob(employer, employee Address, salary uint64, start, end time.Time) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if salary == 0 {
		return "", errors.New("salary zero")
	}
	if !start.Before(end) {
		return "", errors.New("start must be before end")
	}
	id := fmt.Sprintf("emp:%d", r.nextID)
	r.nextID++
	c := EmploymentContract{ID: id, Employer: employer, Employee: employee, SalaryPerHour: salary, Start: start, End: end}
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	if err := r.led.SetState([]byte(id), b); err != nil {
		return "", err
	}
	return id, nil
}

// GetJob retrieves a contract by ID.
func (r *EmploymentRegistry) GetJob(id string) (EmploymentContract, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	raw, err := r.led.GetState([]byte(id))
	if err != nil {
		return EmploymentContract{}, false, err
	}
	if len(raw) == 0 {
		return EmploymentContract{}, false, nil
	}
	var c EmploymentContract
	if err := json.Unmarshal(raw, &c); err != nil {
		return EmploymentContract{}, false, err
	}
	return c, true, nil
}

// SignJob marks employer or employee signature.
func (r *EmploymentRegistry) SignJob(id string, signer Address) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	raw, err := r.led.GetState([]byte(id))
	if err != nil {
		return err
	}
	var c EmploymentContract
	if err := json.Unmarshal(raw, &c); err != nil {
		return err
	}
	if c.Paid {
		return errors.New("contract closed")
	}
	if signer == c.Employer {
		c.EmployerSign = true
	} else if signer == c.Employee {
		c.EmployeeSign = true
	} else {
		return errors.New("not a participant")
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return r.led.SetState([]byte(id), b)
}

// RecordWork increments worked hours.
func (r *EmploymentRegistry) RecordWork(id string, hours uint32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	raw, err := r.led.GetState([]byte(id))
	if err != nil {
		return err
	}
	var c EmploymentContract
	if err := json.Unmarshal(raw, &c); err != nil {
		return err
	}
	if c.Paid {
		return errors.New("contract closed")
	}
	c.HoursWorked += hours
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return r.led.SetState([]byte(id), b)
}

// PaySalary transfers salary and closes the contract.
func (r *EmploymentRegistry) PaySalary(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	raw, err := r.led.GetState([]byte(id))
	if err != nil {
		return err
	}
	var c EmploymentContract
	if err := json.Unmarshal(raw, &c); err != nil {
		return err
	}
	if c.Paid {
		return errors.New("already paid")
	}
	total := uint64(c.HoursWorked) * c.SalaryPerHour
	if err := r.led.Transfer(c.Employer, c.Employee, total); err != nil {
		return err
	}
	if et := employmentToken(); et != nil {
		if err := et.PaySalary(id); err != nil {
			return err
		}
	}
	c.Paid = true
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return r.led.SetState([]byte(id), b)
}
