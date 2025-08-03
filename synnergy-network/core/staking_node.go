package core

import (
	"context"
	"sync"
)

// StakingNode combines networking with staking management for PoS consensus.
type StakingNode struct {
	*BaseNode
	ledger *Ledger
	stake  *DAOStaking
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// StakingConfig groups the configuration for a staking node.
type StakingConfig struct {
	Network Config
	Ledger  LedgerConfig
}

// NewStakingNode initialises networking and the staking manager.
func NewStakingNode(cfg *StakingConfig) (*StakingNode, error) {
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
	InitDAOStaking(nil, led)
	base := NewBaseNode(&NodeAdapter{n})
	return &StakingNode{BaseNode: base, ledger: led, stake: StakingManager(), ctx: ctx, cancel: cancel}, nil
}

// Start begins networking services.
func (s *StakingNode) Start() { go s.ListenAndServe() }

// Stop shuts down the node.
func (s *StakingNode) Stop() error {
	s.cancel()
	return s.Close()
}

// Stake locks tokens via the staking manager.
func (s *StakingNode) Stake(addr Address, amount uint64) error {
	return s.stake.Stake(addr, amount)
}

// Unstake releases previously locked tokens.
func (s *StakingNode) Unstake(addr Address, amount uint64) error {
	return s.stake.Unstake(addr, amount)
}

// ProposeBlock broadcasts a new block proposal.
func (s *StakingNode) ProposeBlock(data []byte) error {
	return s.Broadcast("block_proposal", data)
}

// ValidateBlock broadcasts validation results for a block.
func (s *StakingNode) ValidateBlock(data []byte) error {
	return s.Broadcast("block_validate", data)
}

// Status returns a textual status for monitoring.
func (s *StakingNode) Status() string {
	select {
	case <-s.ctx.Done():
		return "stopped"
	default:
		return "running"
	}
}
