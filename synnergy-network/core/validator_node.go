package core

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/sirupsen/logrus"
)

// ValidatorNode bundles networking, ledger access and consensus participation.
// It exposes helper methods used by opcode handlers and CLI tooling.
type ValidatorNode struct {
	net  *Node
	led  *Ledger
	cons *SynnergyConsensus

	mgr       *ValidatorManager
	penalties *StakePenaltyManager

	ctx    context.Context
	cancel context.CancelFunc

	mu     sync.RWMutex
	usePoH bool
	usePoS bool
	usePoW bool
}

// ValidatorNodeConfig aggregates the required configuration sections.
type ValidatorNodeConfig struct {
	Network Config
	Ledger  LedgerConfig
}

// NewValidatorNode initialises networking, ledger and consensus services.
func NewValidatorNode(cfg ValidatorNodeConfig) (*ValidatorNode, error) {
	ctx, cancel := context.WithCancel(context.Background())
	n, err := NewNode(cfg.Network)
	if err != nil {
		cancel()
		return nil, err
	}
	led, err := NewLedger(cfg.Ledger)
	if err != nil {
		cancel()
		_ = n.Close()
		return nil, err
	}
	mgr := NewValidatorManager(led)
	penalties := NewStakePenaltyManager(logrus.StandardLogger(), led)

	cons, err := NewConsensus(logrus.StandardLogger(), led, n, nil, nil, nil)
	if err != nil {
		cancel()
		_ = n.Close()
		return nil, err
	}

	vn := &ValidatorNode{
		net:       n,
		led:       led,
		cons:      cons,
		mgr:       mgr,
		penalties: penalties,
		ctx:       ctx,
		cancel:    cancel,
		usePoH:    true,
		usePoS:    true,
		usePoW:    true,
	}
	return vn, nil
}

// Start launches network and consensus routines.
func (vn *ValidatorNode) Start() {
	vn.mu.Lock()
	defer vn.mu.Unlock()
	go vn.net.ListenAndServe()
	vn.cons.Start(vn.ctx)
}

// Stop gracefully shuts down the node services.
func (vn *ValidatorNode) Stop() error {
	vn.mu.Lock()
	defer vn.mu.Unlock()
	vn.cancel()
	return vn.net.Close()
}

// EnablePoH toggles the Proof of History component.
func (vn *ValidatorNode) EnablePoH(b bool) { vn.mu.Lock(); vn.usePoH = b; vn.mu.Unlock() }

// EnablePoS toggles the Proof of Stake component.
func (vn *ValidatorNode) EnablePoS(b bool) { vn.mu.Lock(); vn.usePoS = b; vn.mu.Unlock() }

// EnablePoW toggles the Proof of Work component.
func (vn *ValidatorNode) EnablePoW(b bool) { vn.mu.Lock(); vn.usePoW = b; vn.mu.Unlock() }

// ValidateTx verifies a transaction using the consensus tx pool.
func (vn *ValidatorNode) ValidateTx(txBytes []byte) error {
	tx, err := DecodeTransaction(txBytes)
	if err != nil {
		return err
	}
	return vn.cons.pool.ValidateTx(tx)
}

// ProposeBlock triggers block proposal according to the enabled mechanisms.
func (vn *ValidatorNode) ProposeBlock() error {
	_, err := vn.cons.ProposeSubBlock()
	return err
}

// VoteBlock records a PoS vote for the given block header hash.
func (vn *ValidatorNode) VoteBlock(hash []byte, sig []byte) error {
	vn.led.RecordPoSVote(hash, sig)
	return nil
}

// DecodeTransaction converts JSON encoded bytes into a Transaction structure.
func DecodeTransaction(b []byte) (*Transaction, error) {
	var tx Transaction
	if err := json.Unmarshal(b, &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}

// Register exposes the underlying validator manager.
func (vn *ValidatorNode) Register(addr Address, stake uint64) error {
	return vn.mgr.Register(addr, stake)
}

// Deregister removes the validator and returns its stake.
func (vn *ValidatorNode) Deregister(addr Address) error {
	return vn.mgr.Deregister(addr)
}

// Stake adds stake to the validator.
func (vn *ValidatorNode) Stake(addr Address, amt uint64) error {
	return vn.mgr.Stake(addr, amt)
}

// Unstake releases stake back to the validator owner.
func (vn *ValidatorNode) Unstake(addr Address, amt uint64) error {
	return vn.mgr.Unstake(addr, amt)
}

// Slash penalises the validator by burning part of its stake.
func (vn *ValidatorNode) Slash(addr Address, amt uint64) error {
	return vn.mgr.Slash(addr, amt)
}

// Info retrieves validator information.
func (vn *ValidatorNode) Info(addr Address) (ValidatorInfo, error) {
	return vn.mgr.Get(addr)
}

// List returns all validators. If activeOnly is true only active ones are listed.
func (vn *ValidatorNode) List(activeOnly bool) ([]ValidatorInfo, error) {
	return vn.mgr.List(activeOnly)
}

// IsValidator checks whether the address is an active validator.
func (vn *ValidatorNode) IsValidator(addr Address) bool { return vn.mgr.IsValidator(addr) }
