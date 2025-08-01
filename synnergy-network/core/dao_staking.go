package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
)

// DAOStaking manages token staking for governance participation.
// Balances are persisted in the ledger key/value store under the
// prefix "dao:stake:". The total locked amount is tracked separately
// for quick access by consensus or reward modules.

type DAOStaking struct {
	logger *log.Logger
	ledger StateRW
	mu     sync.Mutex
}

var (
	stakingOnce sync.Once
	stakingMgr  *DAOStaking
)

// InitDAOStaking initialises the global staking manager. It must be
// called before using any staking operations.
func InitDAOStaking(lg *log.Logger, led StateRW) {
	stakingOnce.Do(func() {
		stakingMgr = &DAOStaking{logger: lg, ledger: led}
	})
}

// StakingManager returns the singleton staking engine.
func StakingManager() *DAOStaking { return stakingMgr }

const (
	stakePrefix = "dao:stake:"
	totalKey    = "dao:stake:total"
)

// Stake locks the given amount of the base coin from addr.
func (s *DAOStaking) Stake(addr Address, amt uint64) error {
	if s == nil || s.ledger == nil {
		return errors.New("staking not initialised")
	}
	if amt == 0 {
		return errors.New("zero amount")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ledger.Transfer(addr, AddressZero, amt); err != nil {
		return err
	}

	bal := s.stakedOf(addr)
	bal += amt
	b, _ := json.Marshal(bal)
	if err := s.ledger.SetState([]byte(stakePrefix+addr.String()), b); err != nil {
		return err
	}

	tot := s.totalLocked() + amt
	tb, _ := json.Marshal(tot)
	if err := s.ledger.SetState([]byte(totalKey), tb); err != nil {
		return err
	}
	s.logger.Printf("stake %d from %s", amt, addr.Short())
	return nil
}

// Unstake releases previously locked coins back to addr.
func (s *DAOStaking) Unstake(addr Address, amt uint64) error {
	if s == nil || s.ledger == nil {
		return errors.New("staking not initialised")
	}
	if amt == 0 {
		return errors.New("zero amount")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	bal := s.stakedOf(addr)
	if bal < amt {
		return fmt.Errorf("insufficient staked balance")
	}
	bal -= amt
	b, _ := json.Marshal(bal)
	if err := s.ledger.SetState([]byte(stakePrefix+addr.String()), b); err != nil {
		return err
	}

	tot := s.totalLocked() - amt
	tb, _ := json.Marshal(tot)
	if err := s.ledger.SetState([]byte(totalKey), tb); err != nil {
		return err
	}

	if err := s.ledger.Transfer(AddressZero, addr, amt); err != nil {
		return err
	}
	s.logger.Printf("unstake %d to %s", amt, addr.Short())
	return nil
}

// StakedOf returns the current staked amount for addr.
func (s *DAOStaking) StakedOf(addr Address) uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stakedOf(addr)
}

// TotalStaked returns the total tokens locked across all accounts.
func (s *DAOStaking) TotalStaked() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.totalLocked()
}

func (s *DAOStaking) stakedOf(addr Address) uint64 {
	raw, _ := s.ledger.GetState([]byte(stakePrefix + addr.String()))
	if len(raw) == 0 {
		return 0
	}
	var bal uint64
	_ = json.Unmarshal(raw, &bal)
	return bal
}

func (s *DAOStaking) totalLocked() uint64 {
	raw, _ := s.ledger.GetState([]byte(totalKey))
	if len(raw) == 0 {
		return 0
	}
	var tot uint64
	_ = json.Unmarshal(raw, &tot)
	return tot
}
