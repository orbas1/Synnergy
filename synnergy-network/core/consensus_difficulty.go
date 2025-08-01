package core

import (
	"errors"
	"math/big"
)

// ConsensusStatus exposes high level consensus metrics such as the current
// PoW difficulty and ledger heights. It is used by CLI commands and smart
// contracts to query the state of the consensus engine.
type ConsensusStatus struct {
	Difficulty     *big.Int
	BlockHeight    uint64
	SubBlockHeight uint64
}

// Status returns the current consensus status. The returned big.Int is a copy
// and can be mutated by the caller without affecting consensus state.
func (sc *SynnergyConsensus) Status() ConsensusStatus {
	return ConsensusStatus{
		Difficulty:     sc.getDifficulty(),
		BlockHeight:    sc.ledger.LastBlockHeight(),
		SubBlockHeight: sc.ledger.LastSubBlockHeight(),
	}
}

// SetDifficulty manually adjusts the proof-of-work difficulty. It is primarily
// intended for testing and network administration. A non-nil, positive value is
// required.
func (sc *SynnergyConsensus) SetDifficulty(diff *big.Int) error {
	if diff == nil || diff.Sign() <= 0 {
		return errors.New("invalid difficulty")
	}
	sc.mu.Lock()
	sc.curDifficulty = new(big.Int).Set(diff)
	sc.mu.Unlock()
	return nil
}
