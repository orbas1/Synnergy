package internal

import (
	"errors"
	"sync"

	log "github.com/sirupsen/logrus"
	core "synnergy-network/core"
)

// CharityPoolManager extends core.CharityPool with enterprise oriented
// management utilities. It exposes donation and withdrawal helpers as well
// as real time balance introspection used by external services.
type CharityPoolManager struct {
	cp     *core.CharityPool
	ledger *core.Ledger
	logger *log.Logger
	mu     sync.Mutex
}

var (
	// ErrAmountZero is returned when the supplied token amount is zero.
	ErrAmountZero = errors.New("amount must be greater than zero")
	// errInsufficientBalance is returned when an account lacks the funds
	// required to complete an operation.
	errInsufficientBalance = errors.New("insufficient balance")
)

// CharityBalances represents the current fund distribution between the
// donation pool and the internal charity wallet.
type CharityBalances struct {
	Pool     uint64 `json:"pool"`
	Internal uint64 `json:"internal"`
}

// NewCharityPoolManager wires a CharityPool instance with the ledger so
// management operations can atomically move funds. The logger may be nil.
func NewCharityPoolManager(lg *log.Logger, cp *core.CharityPool, led *core.Ledger) *CharityPoolManager {
	if lg == nil {
		lg = log.StandardLogger()
	}
	return &CharityPoolManager{cp: cp, ledger: led, logger: lg}
}

// Donate transfers tokens from the given address directly into the charity
// pool. This is distinct from the automatic gas fee contribution and allows
// users or smart contracts to donate voluntarily.
func (m *CharityPoolManager) Donate(from core.Address, amount uint64) error {
	if amount == 0 {
		return ErrAmountZero
	}
	m.logger.Printf("donation %d from %s", amount, from.Short())
	return m.ledger.Transfer(from, core.CharityPoolAccount, amount)
}

// WithdrawInternal moves funds from the InternalCharityAccount to a target
// address. Only callable by governance or authorised actors at the application
// layer.
func (m *CharityPoolManager) WithdrawInternal(to core.Address, amount uint64) error {
	if amount == 0 {
		return ErrAmountZero
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	bal := m.ledger.BalanceOf(core.InternalCharityAccount)
	if amount > bal {
		return errInsufficientBalance
	}
	m.logger.Printf("internal withdrawal %d to %s", amount, to.Short())
	return m.ledger.Transfer(core.InternalCharityAccount, to, amount)
}

// Balances returns the current pool and internal account balances.
// These values are pulled from the ledger atomically and packaged for
// monitoring dashboards and CLI inspection.
func (m *CharityPoolManager) Balances() CharityBalances {
	return CharityBalances{
		Pool:     m.ledger.BalanceOf(core.CharityPoolAccount),
		Internal: m.ledger.BalanceOf(core.InternalCharityAccount),
	}
}
