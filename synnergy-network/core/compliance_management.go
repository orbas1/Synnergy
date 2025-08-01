package core

import (
	"errors"
	"fmt"
	"sync"
)

// ComplianceManager coordinates account suspensions and whitelists.
type ComplianceManager struct {
	mu     sync.RWMutex
	ledger *Ledger
}

var (
	cmOnce sync.Once
	cm     *ComplianceManager
)

// InitComplianceManager initialises the global manager instance.
func InitComplianceManager(led *Ledger) {
	cmOnce.Do(func() {
		cm = &ComplianceManager{ledger: led}
	})
}

// ComplianceMgmt returns the singleton ComplianceManager.
func ComplianceMgmt() *ComplianceManager { return cm }

// SuspendAccount marks an address as suspended in the ledger state.
func (m *ComplianceManager) SuspendAccount(addr Address) error {
	return m.ledger.SetState(m.suspendKey(addr), []byte{1})
}

// ResumeAccount lifts a previously applied suspension.
func (m *ComplianceManager) ResumeAccount(addr Address) error {
	return m.ledger.DeleteState(m.suspendKey(addr))
}

// IsSuspended returns true if the address is currently suspended.
func (m *ComplianceManager) IsSuspended(addr Address) bool {
	b, _ := m.ledger.GetState(m.suspendKey(addr))
	return len(b) > 0
}

// WhitelistAccount exempts an address from suspension checks.
func (m *ComplianceManager) WhitelistAccount(addr Address) error {
	return m.ledger.SetState(m.whitelistKey(addr), []byte{1})
}

// RemoveWhitelist deletes a whitelist entry.
func (m *ComplianceManager) RemoveWhitelist(addr Address) error {
	return m.ledger.DeleteState(m.whitelistKey(addr))
}

// IsWhitelisted reports whether the address is whitelisted.
func (m *ComplianceManager) IsWhitelisted(addr Address) bool {
	b, _ := m.ledger.GetState(m.whitelistKey(addr))
	return len(b) > 0
}

// ReviewTransaction ensures both parties are allowed to transact.
func (m *ComplianceManager) ReviewTransaction(tx *Transaction) error {
	if tx == nil {
		return errors.New("nil tx")
	}
	if m.IsSuspended(tx.From) && !m.IsWhitelisted(tx.From) {
		return fmt.Errorf("sender %x suspended", tx.From)
	}
	if m.IsSuspended(tx.To) && !m.IsWhitelisted(tx.To) {
		return fmt.Errorf("recipient %x suspended", tx.To)
	}
	return nil
}

func (m *ComplianceManager) suspendKey(addr Address) []byte {
	return append([]byte("susp:"), addr[:]...)
}

func (m *ComplianceManager) whitelistKey(addr Address) []byte {
	return append([]byte("white:"), addr[:]...)
}
