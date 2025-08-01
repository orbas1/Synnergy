package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// SimplePlasmaDeposit represents a deposit into the Plasma chain.
type SimplePlasmaDeposit struct {
	Nonce     uint64  `json:"nonce"`
	From      Address `json:"from"`
	Amount    uint64  `json:"amount"`
	Timestamp int64   `json:"ts"`
}

// PlasmaExit records a finalised withdrawal from the Plasma chain.
type SimplePlasmaExit struct {
	Nonce     uint64  `json:"nonce"`
	To        Address `json:"to"`
	Proof     []byte  `json:"proof"`
	Timestamp int64   `json:"ts"`
}

// PlasmaCoordinator manages deposits and exits for a simple Plasma child chain.
type PlasmaCoordinator struct {
	Ledger StateRW
	mu     sync.Mutex
	nonce  uint64
}

var (
	plasmaOnce sync.Once
	plasma     *PlasmaCoordinator
)

// InitPlasma initialises the global Plasma coordinator with the given ledger.
func InitPlasma(led StateRW) {
	plasmaOnce.Do(func() {
		plasma = &PlasmaCoordinator{Ledger: led}
	})
}

// Plasma returns the global Plasma coordinator instance.
func Plasma() *PlasmaCoordinator { return plasma }

// Deposit locks funds on L1 and records a Plasma deposit.
func (pc *PlasmaCoordinator) Deposit(from Address, amount uint64) (uint64, error) {
	if pc == nil {
		return 0, errors.New("plasma not initialised")
	}
	if amount == 0 {
		return 0, errors.New("amount zero")
	}
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.nonce++
	n := pc.nonce
	dep := SimplePlasmaDeposit{Nonce: n, From: from, Amount: amount, Timestamp: time.Now().Unix()}
	raw, _ := json.Marshal(dep)
	_ = pc.Ledger.SetState(pc.depKey(n), raw)
	return n, nil
}

// Withdraw finalises an exit by deleting the deposit record.
func (pc *PlasmaCoordinator) Withdraw(exit SimplePlasmaExit) error {
	if pc == nil {
		return errors.New("plasma not initialised")
	}
	pc.mu.Lock()
	defer pc.mu.Unlock()
	key := pc.depKey(exit.Nonce)
	exists, _ := pc.Ledger.HasState(key)
	if !exists {
		return errors.New("deposit not found")
	}
	_ = pc.Ledger.DeleteState(key)
	exit.Timestamp = time.Now().Unix()
	raw, _ := json.Marshal(exit)
	_ = pc.Ledger.SetState(pc.exitKey(exit.Nonce), raw)
	return nil
}

func (pc *PlasmaCoordinator) depKey(n uint64) []byte {
	return []byte(fmt.Sprintf("plasma:dep:%d", n))
}

func (pc *PlasmaCoordinator) exitKey(n uint64) []byte {
	return []byte(fmt.Sprintf("plasma:exit:%d", n))
}
