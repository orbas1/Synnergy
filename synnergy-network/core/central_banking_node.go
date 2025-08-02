package core

import (
	"math/big"
	"sync"

	"github.com/sirupsen/logrus"
)

// CentralBankingNode embeds a networking node and exposes hooks for monetary
// policy and settlement operations used by central banks.
type CentralBankingNode struct {
	*BaseNode
	ledger *Ledger
	cons   *SynnergyConsensus

	mu                 sync.RWMutex
	interestRate       float64
	reserveRequirement float64
}

// NewCentralBankingNode instantiates the node with the provided network and
// ledger configurations. Consensus may be nil if the node only performs
// regulatory functions.
func NewCentralBankingNode(netCfg Config, ledCfg LedgerConfig, cons *SynnergyConsensus) (*CentralBankingNode, error) {
	n, err := NewNode(netCfg)
	if err != nil {
		return nil, err
	}
	led, err := NewLedger(ledCfg)
	if err != nil {
		_ = n.Close()
		return nil, err
	}
	base := NewBaseNode(&NodeAdapter{n})
	cb := &CentralBankingNode{
		BaseNode:           base,
		ledger:             led,
		cons:               cons,
		interestRate:       0,
		reserveRequirement: 0,
	}
	return cb, nil
}

// Start begins network services.
func (cb *CentralBankingNode) Start() {
	go cb.ListenAndServe()
}

// Stop shuts down the networking layer.
func (cb *CentralBankingNode) Stop() error {
	return cb.Close()
}

// SetInterestRate updates the reference interest rate used by the monetary
// policy tools.
func (cb *CentralBankingNode) SetInterestRate(r float64) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.interestRate = r
	logrus.Infof("central bank interest rate set to %.2f", r)
	return nil
}

// InterestRate returns the current interest rate.
func (cb *CentralBankingNode) InterestRate() float64 {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.interestRate
}

// SetReserveRequirement updates the reserve ratio required for institutions.
func (cb *CentralBankingNode) SetReserveRequirement(r float64) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.reserveRequirement = r
	logrus.Infof("reserve requirement set to %.2f", r)
	return nil
}

// ReserveRequirement returns the active reserve ratio.
func (cb *CentralBankingNode) ReserveRequirement() float64 {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.reserveRequirement
}

// IssueDigitalCurrency mints new tokens directly into the ledger under control
// of the central bank.
func (cb *CentralBankingNode) IssueDigitalCurrency(addr Address, amount uint64) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	amt := new(big.Int).SetUint64(amount)
	cb.ledger.MintBig(addr[:], amt)
	logrus.Infof("issued %d units to %s", amount, addr.Short())
	return nil
}

// RecordSettlement stores a transaction in the ledger for real-time settlement.
func (cb *CentralBankingNode) RecordSettlement(tx *Transaction) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	h := tx.Hash().Hex()
	cb.ledger.TxPool[h] = tx
	logrus.WithField("tx", h).Info("settlement recorded")
	return nil
}
