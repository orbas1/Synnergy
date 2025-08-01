package core

import (
	"fmt"
	"sync"
)

// AccountManager provides helper operations for creating accounts and
// manipulating their coin balances. It wraps a Ledger instance and
// performs thread-safe updates.
type AccountManager struct {
	ledger *Ledger
	mu     sync.RWMutex
}

// NewAccountManager constructs a manager bound to the given ledger.
func NewAccountManager(l *Ledger) *AccountManager {
	return &AccountManager{ledger: l}
}

// CreateAccount initialises a zero balance entry for addr. An error is
// returned if the account already exists or the ledger is nil.
func (am *AccountManager) CreateAccount(addr Address) error {
	if am.ledger == nil {
		return fmt.Errorf("account manager: nil ledger")
	}
	am.mu.Lock()
	defer am.mu.Unlock()
	key := addr.String()
	if _, ok := am.ledger.TokenBalances[key]; ok {
		return fmt.Errorf("account %s exists", key)
	}
	am.ledger.TokenBalances[key] = 0
	return nil
}

// DeleteAccount removes addr from the ledger balance map.
func (am *AccountManager) DeleteAccount(addr Address) error {
	if am.ledger == nil {
		return fmt.Errorf("account manager: nil ledger")
	}
	am.mu.Lock()
	defer am.mu.Unlock()
	key := addr.String()
	if _, ok := am.ledger.TokenBalances[key]; !ok {
		return fmt.Errorf("account %s not found", key)
	}
	delete(am.ledger.TokenBalances, key)
	return nil
}

// Balance returns the current coin balance for addr.
func (am *AccountManager) Balance(addr Address) (uint64, error) {
	if am.ledger == nil {
		return 0, fmt.Errorf("account manager: nil ledger")
	}
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.ledger.TokenBalances[addr.String()], nil
}

// Transfer moves amt coins from src to dst, verifying sufficient funds.
func (am *AccountManager) Transfer(src, dst Address, amt uint64) error {
	if am.ledger == nil {
		return fmt.Errorf("account manager: nil ledger")
	}
	if amt == 0 {
		return fmt.Errorf("transfer amount must be positive")
	}
	am.mu.Lock()
	defer am.mu.Unlock()
	if am.ledger.TokenBalances[src.String()] < amt {
		return fmt.Errorf("insufficient balance")
	}
	am.ledger.TokenBalances[src.String()] -= amt
	am.ledger.TokenBalances[dst.String()] += amt
	return nil
}
