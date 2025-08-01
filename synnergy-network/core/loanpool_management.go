package core

import (
	"encoding/json"
	"sync"
	"time"
)

// LoanPoolManager provides administrative helpers around LoanPool.
// It exposes basic pause/resume semantics and a snapshot of treasury stats.
type LoanPoolManager struct {
	pool   *LoanPool
	paused bool
	mu     sync.RWMutex
}

type LoanPoolStats struct {
	Treasury uint64
	Active   int
	Passed   int
	Rejected int
}

// NewLoanPoolManager wraps an existing LoanPool instance.
func NewLoanPoolManager(lp *LoanPool) *LoanPoolManager {
	return &LoanPoolManager{pool: lp}
}

// Pause prevents new proposals from being submitted.
func (m *LoanPoolManager) Pause() { m.mu.Lock(); m.paused = true; m.mu.Unlock() }

// Resume re-enables proposal submission.
func (m *LoanPoolManager) Resume() { m.mu.Lock(); m.paused = false; m.mu.Unlock() }

// IsPaused returns true if the loan pool is currently paused.
func (m *LoanPoolManager) IsPaused() bool { m.mu.RLock(); v := m.paused; m.mu.RUnlock(); return v }

// Stats returns high level metrics about the loan pool.
func (m *LoanPoolManager) Stats() LoanPoolStats {
	m.pool.mu.Lock()
	defer m.pool.mu.Unlock()

	iter := m.pool.ledger.PrefixIterator([]byte("loanpool:proposal:"))
	var s LoanPoolStats
	s.Treasury = m.pool.ledger.BalanceOf(LoanPoolAccount)
	for iter.Next() {
		var p Proposal
		_ = json.Unmarshal(iter.Value(), &p)
		switch p.Status {
		case Active:
			s.Active++
		case Passed:
			s.Passed++
		case Rejected:
			s.Rejected++
		}
	}
	return s
}

// Tick proxies to the underlying pool but aborts if paused.
func (m *LoanPoolManager) Tick(now time.Time) {
	if m.IsPaused() {
		return
	}
	m.pool.Tick(now)
}
