package core

import (
	"context"
	"sync"

	Nodes "synnergy-network/core/Nodes"
)

// StakingNode combines networking with staking management for PoS consensus.
type StakingNode struct {
	net    *Node
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
	return &StakingNode{net: n, ledger: led, stake: StakingManager(), ctx: ctx, cancel: cancel}, nil
}

// Start begins networking services.
func (s *StakingNode) Start() { s.net.ListenAndServe() }

// Stop shuts down the node.
func (s *StakingNode) Stop() error {
	s.cancel()
	return s.net.Close()
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
	return s.net.Broadcast("block_proposal", data)
}

// ValidateBlock broadcasts validation results for a block.
func (s *StakingNode) ValidateBlock(data []byte) error {
	return s.net.Broadcast("block_validate", data)
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

// DialSeed proxies to the underlying network node.
func (s *StakingNode) DialSeed(peers []string) error { return s.net.DialSeed(peers) }

// Broadcast proxies to the underlying network node.
func (s *StakingNode) Broadcast(topic string, data []byte) error { return s.net.Broadcast(topic, data) }

// Subscribe proxies to the underlying network node.
func (s *StakingNode) Subscribe(topic string) (<-chan []byte, error) { return s.net.Subscribe(topic) }

// ListenAndServe proxies to the underlying network node.
func (s *StakingNode) ListenAndServe() { s.net.ListenAndServe() }

// Close terminates the node.
func (s *StakingNode) Close() error { return s.net.Close() }

// Peers lists known peers.
func (s *StakingNode) Peers() []string {
	peers := s.net.Peers()
	out := make([]string, len(peers))
	for i, p := range peers {
		out[i] = string(p.ID)
	}
	return out
}

var _ Nodes.StakingNodeInterface = (*StakingNode)(nil)
