package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
)

// StakePenaltyManager provides helper methods for adjusting validator stake
// and recording penalties for misbehaviour. It stores data in the ledger using
// the generic StateRW interface so it can operate on any compatible backend.
//
// Stake amounts are tracked under the key "stake:<addr>" and penalty points
// under "penalty:<addr>". Values are encoded as big-endian integers.
//
// The manager is concurrency safe and intended to be used by consensus
// components or administrative tooling.
type StakePenaltyManager struct {
	led    StateRW
	logger *log.Logger
	mu     sync.RWMutex
}

// NewStakePenaltyManager constructs a new manager with the provided logger and
// StateRW implementation.
func NewStakePenaltyManager(lg *log.Logger, led StateRW) *StakePenaltyManager {
	return &StakePenaltyManager{logger: lg, led: led}
}

// AdjustStake increases or decreases the recorded stake for an address. A
// negative delta is allowed so long as the resulting stake does not go below
// zero.
func (spm *StakePenaltyManager) AdjustStake(addr Address, delta int64) error {
	spm.mu.Lock()
	defer spm.mu.Unlock()
	key := stakeKey(addr)
	curRaw, err := spm.led.GetState(key)
	if err != nil {
		return err
	}
	var cur int64
	if len(curRaw) != 0 {
		cur = int64(binary.BigEndian.Uint64(curRaw))
	}
	next := cur + delta
	if next < 0 {
		return fmt.Errorf("insufficient stake")
	}
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(next))
	return spm.led.SetState(key, buf)
}

// StakeOf returns the currently recorded stake for the address.
func (spm *StakePenaltyManager) StakeOf(addr Address) uint64 {
	spm.mu.RLock()
	defer spm.mu.RUnlock()
	raw, err := spm.led.GetState(stakeKey(addr))
	if err != nil || len(raw) == 0 {
		return 0
	}
	return binary.BigEndian.Uint64(raw)
}

// Penalize adds penalty points for a validator and logs the reason. Penalties
// accumulate over time and may be used by consensus to slash or deactivate
// misbehaving nodes.
func (spm *StakePenaltyManager) Penalize(addr Address, points uint32, reason string) error {
	spm.mu.Lock()
	defer spm.mu.Unlock()
	key := penaltyKey(addr)
	raw, err := spm.led.GetState(key)
	if err != nil {
		return err
	}
	var cur uint32
	if len(raw) != 0 {
		cur = binary.BigEndian.Uint32(raw)
	}
	cur += points
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, cur)
	if err := spm.led.SetState(key, buf); err != nil {
		return err
	}
	spm.logger.WithFields(log.Fields{"addr": addr, "points": points, "reason": reason}).Warn("validator penalized")
	return nil
}

// PenaltyOf returns the accumulated penalty points for the validator.
func (spm *StakePenaltyManager) PenaltyOf(addr Address) uint32 {
	spm.mu.RLock()
	defer spm.mu.RUnlock()
	raw, err := spm.led.GetState(penaltyKey(addr))
	if err != nil || len(raw) == 0 {
		return 0
	}
	return binary.BigEndian.Uint32(raw)
}

func stakeKey(addr Address) []byte   { return []byte("stake:" + addr.Hex()) }
func penaltyKey(addr Address) []byte { return []byte("penalty:" + addr.Hex()) }

// SlashStake reduces the recorded stake for an address by the given fraction
// (e.g. 0.25 slashes 25%). The slashed amount is returned. An error is
// returned if no stake is recorded or the ledger update fails.
func (spm *StakePenaltyManager) SlashStake(addr Address, fraction float64) (uint64, error) {
	spm.mu.Lock()
	defer spm.mu.Unlock()

	if fraction <= 0 || fraction > 1 {
		return 0, fmt.Errorf("fraction must be within (0,1]")
	}

	key := stakeKey(addr)
	raw, err := spm.led.GetState(key)
	if err != nil {
		return 0, err
	}
	if len(raw) == 0 {
		return 0, errors.New("no stake recorded")
	}

	cur := binary.BigEndian.Uint64(raw)
	slash := uint64(float64(cur) * fraction)
	if slash > cur {
		slash = cur
	}
	next := cur - slash
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, next)
	if err := spm.led.SetState(key, buf); err != nil {
		return 0, err
	}
	spm.logger.WithFields(log.Fields{"addr": addr, "slashed": slash}).Warn("stake slashed")
	return slash, nil
}

// ResetPenalty clears accumulated penalty points for the address and records the action.
func (spm *StakePenaltyManager) ResetPenalty(addr Address) error {
	spm.mu.Lock()
	defer spm.mu.Unlock()
	if err := spm.led.DeleteState(penaltyKey(addr)); err != nil {
		return err
	}
	if spm.logger != nil {
		spm.logger.WithField("addr", addr).Info("penalties reset")
	}
	return nil
}
