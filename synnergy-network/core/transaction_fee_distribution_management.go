package core

import "sync"

// TxFeeManager manages transaction fee collection and distribution.
type TxFeeManager struct {
	ledger *Ledger
	mu     sync.Mutex
	fees   uint64
}

// FeeCollectorAccount temporarily holds deducted fees before distribution.
var FeeCollectorAccount Address

func init() {
	FeeCollectorAccount = ModuleAddress("fee_collector")
}

// NewTxFeeManager returns a new fee manager bound to the given ledger.
func NewTxFeeManager(l *Ledger) *TxFeeManager {
	return &TxFeeManager{ledger: l}
}

// Collect records a fee from the sender and allocates the charity portion.
func (m *TxFeeManager) Collect(from Address, amount uint64) {
	if amount == 0 {
		return
	}
	// move tokens to the fee collector account
	_ = m.ledger.Transfer(from, FeeCollectorAccount, amount)

	// 5%% of each fee goes to the charity pool
	charity := amount * 5 / 100
	if charity > 0 {
		_ = m.ledger.Transfer(FeeCollectorAccount, CharityPoolAccount, charity)
		amount -= charity
	}

	m.mu.Lock()
	m.fees += amount
	m.mu.Unlock()
}

// Distribute splits all collected fees between miner, validators and the loan pool.
func (m *TxFeeManager) Distribute(miner Address, validators []Address) {
	m.mu.Lock()
	total := m.fees
	m.fees = 0
	m.mu.Unlock()

	if total == 0 {
		return
	}

	minerShare := total * 30 / 100
	stakerShare := total * 30 / 100
	loanShare := total - minerShare - stakerShare

	_ = m.ledger.Transfer(FeeCollectorAccount, miner, minerShare)

	if len(validators) > 0 {
		per := stakerShare / uint64(len(validators))
		for _, v := range validators {
			_ = m.ledger.Transfer(FeeCollectorAccount, v, per)
		}
	} else {
		loanShare += stakerShare
	}

	_ = m.ledger.Transfer(FeeCollectorAccount, LoanPoolAccount, loanShare)
}

// Pending returns currently collected but undistributed fees.
func (m *TxFeeManager) Pending() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.fees
}

var (
	feeMgrOnce sync.Once
	feeMgr     *TxFeeManager
)

// InitTxFeeManager initialises the global fee manager.
func InitTxFeeManager(l *Ledger) {
	feeMgrOnce.Do(func() { feeMgr = NewTxFeeManager(l) })
}

// CurrentTxFeeManager returns the global fee manager instance.
func CurrentTxFeeManager() *TxFeeManager { return feeMgr }
