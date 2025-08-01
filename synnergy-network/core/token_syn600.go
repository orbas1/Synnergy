package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// SYN600Token implements reward token mechanics including staking and
// engagement tracking. It embeds the core.BaseToken for standard token
// functionality and extends it with reward specific helpers.

type StakeRecord struct {
	Amount uint64 `json:"amount"`
	Unlock int64  `json:"unlock"`
}

// SYN600Token is the concrete token type used for reward distribution.
type SYN600Token struct {
	*BaseToken
	ledger StateRW
	mu     sync.Mutex
}

const (
	stakePrefix  = "syn600:stake:"
	engagePrefix = "syn600:eng:"
)

func NewSYN600Token(bt *BaseToken, led StateRW) *SYN600Token {
	return &SYN600Token{BaseToken: bt, ledger: led}
}

func (t *SYN600Token) stakeKey(a Address) []byte  { return []byte(stakePrefix + a.String()) }
func (t *SYN600Token) engageKey(a Address) []byte { return []byte(engagePrefix + a.String()) }

// Stake locks tokens for a period and records the stake in the ledger state.
func (t *SYN600Token) Stake(addr Address, amt uint64, dur time.Duration) error {
	if amt == 0 {
		return fmt.Errorf("amount must be >0")
	}
	if err := t.Transfer(addr, AddressZero, amt); err != nil {
		return err
	}
	rec := StakeRecord{Amount: amt, Unlock: time.Now().Add(dur).Unix()}
	b, _ := json.Marshal(rec)
	return t.ledger.SetState(t.stakeKey(addr), b)
}

// Unstake releases locked tokens once the lock duration has expired.
func (t *SYN600Token) Unstake(addr Address) error {
	raw, err := t.ledger.GetState(t.stakeKey(addr))
	if err != nil || len(raw) == 0 {
		return fmt.Errorf("no stake for address")
	}
	var rec StakeRecord
	_ = json.Unmarshal(raw, &rec)
	if time.Now().Unix() < rec.Unlock {
		return fmt.Errorf("stake still locked")
	}
	_ = t.ledger.DeleteState(t.stakeKey(addr))
	return t.Transfer(AddressZero, addr, rec.Amount)
}

// AddEngagement increases the engagement score for the given user.
func (t *SYN600Token) AddEngagement(addr Address, pts uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	raw, _ := t.ledger.GetState(t.engageKey(addr))
	var score uint64
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &score)
	}
	score += pts
	b, _ := json.Marshal(score)
	return t.ledger.SetState(t.engageKey(addr), b)
}

// EngagementOf returns the recorded engagement score for addr.
func (t *SYN600Token) EngagementOf(addr Address) uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	raw, _ := t.ledger.GetState(t.engageKey(addr))
	if len(raw) == 0 {
		return 0
	}
	var score uint64
	_ = json.Unmarshal(raw, &score)
	return score
}

// DistributeStakingRewards mints a percentage of the staked amount as reward.
// rate is interpreted as parts-per-hundred (e.g. 5 = 5%).
func (t *SYN600Token) DistributeStakingRewards(rate uint64) error {
	it := t.ledger.PrefixIterator([]byte(stakePrefix))
	for it.Next() {
		key := it.Key()[len(stakePrefix):]
		b, err := hex.DecodeString(string(key))
		if err != nil || len(b) != 20 {
			continue
		}
		var addr Address
		copy(addr[:], b)
		var rec StakeRecord
		if err := json.Unmarshal(it.Value(), &rec); err != nil {
			continue
		}
		reward := rec.Amount * rate / 100
		_ = t.Mint(addr, reward)
	}
	return nil
}
