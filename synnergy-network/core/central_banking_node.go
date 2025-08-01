package core

import (
	"math/big"
	"sync"

	"github.com/sirupsen/logrus"
	Nodes "synnergy-network/core/Nodes"
)

// CentralBankingNode implements Nodes.CentralBankingNode. It embeds a
// networking node and exposes hooks for monetary policy and settlement
// operations used by central banks.
type CentralBankingNode struct {
	net    *Node
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
	cb := &CentralBankingNode{
		net:                n,
		ledger:             led,
		cons:               cons,
		interestRate:       0,
		reserveRequirement: 0,
	}
	return cb, nil
}

// Start begins network services.
func (cb *CentralBankingNode) Start() {
	go cb.net.ListenAndServe()
}

// Stop shuts down the networking layer.
func (cb *CentralBankingNode) Stop() error {
	return cb.net.Close()
}

// DialSeed proxies to the underlying node.
func (cb *CentralBankingNode) DialSeed(peers []string) error             { return cb.net.DialSeed(peers) }
func (cb *CentralBankingNode) Broadcast(t string, d []byte) error        { return cb.net.Broadcast(t, d) }
func (cb *CentralBankingNode) Subscribe(t string) (<-chan []byte, error) { return cb.net.Subscribe(t) }
func (cb *CentralBankingNode) ListenAndServe()                           { cb.net.ListenAndServe() }
func (cb *CentralBankingNode) Close() error                              { return cb.net.Close() }
func (cb *CentralBankingNode) Peers() []string                           { return cb.net.Peers() }

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

// compile-time check
var _ Nodes.CentralBankingNode = (*CentralBankingNode)(nil)
