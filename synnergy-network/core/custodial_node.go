package core

import (
	"encoding/json"
	"errors"
	"sync"
)

// CustodialConfig bundles network and ledger configuration for CustodialNode.
type CustodialConfig struct {
	Network Config
	Ledger  LedgerConfig
}

// errInsufficientBalance is returned when a withdrawal or transfer exceeds the
// stored amount for an account.
var errInsufficientBalance = errors.New("insufficient balance")

// CustodialNode provides secure asset custody and management services.
type CustodialNode struct {
	*BaseNode
	ledger *Ledger
	store  map[Address]map[TokenID]uint64
	mu     sync.RWMutex
}

// NewCustodialNode initialises networking and ledger access for custody services.
func NewCustodialNode(cfg CustodialConfig) (*CustodialNode, error) {
	n, err := NewNode(cfg.Network)
	if err != nil {
		return nil, err
	}
	led, err := NewLedger(cfg.Ledger)
	if err != nil {
		_ = n.Close()
		return nil, err
	}
	base := NewBaseNode(&NodeAdapter{n})
	c := &CustodialNode{BaseNode: base, ledger: led, store: make(map[Address]map[TokenID]uint64)}
	return c, nil
}

// Start begins the underlying network services.
func (c *CustodialNode) Start() { go c.ListenAndServe() }

// Stop closes network connections and flushes state.
func (c *CustodialNode) Stop() error { return c.Close() }

// Register prepares internal storage for the account.
func (c *CustodialNode) Register(addr Address) {
	c.mu.Lock()
	if _, ok := c.store[addr]; !ok {
		c.store[addr] = make(map[TokenID]uint64)
	}
	c.mu.Unlock()
}

// Deposit credits the account and mints tokens on the ledger for tracking.
func (c *CustodialNode) Deposit(addr Address, token TokenID, amount uint64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.store[addr]; !ok {
		c.store[addr] = make(map[TokenID]uint64)
	}
	c.store[addr][token] += amount
	return c.ledger.Mint(addr, amount)
}

// Withdraw debits the account and burns tokens from the ledger.
func (c *CustodialNode) Withdraw(addr Address, token TokenID, amount uint64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	bal := c.store[addr][token]
	if bal < amount {
		return errInsufficientBalance
	}
	c.store[addr][token] -= amount
	return c.ledger.Burn(addr, amount)
}

// Transfer moves assets between custodial accounts and updates the ledger.
func (c *CustodialNode) Transfer(from, to Address, token TokenID, amount uint64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.store[from][token] < amount {
		return errInsufficientBalance
	}
	if _, ok := c.store[to]; !ok {
		c.store[to] = make(map[TokenID]uint64)
	}
	c.store[from][token] -= amount
	c.store[to][token] += amount
	return c.ledger.Transfer(from, to, amount)
}

// BalanceOf returns the current balance for an account.
func (c *CustodialNode) BalanceOf(addr Address, token TokenID) uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.store[addr][token]
}

// Audit returns a JSON representation of all holdings for external reporting.
func (c *CustodialNode) Audit() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return json.Marshal(c.store)
}
