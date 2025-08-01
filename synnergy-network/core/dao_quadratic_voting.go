package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"
)

// QuadraticVoteRecord stores an individual quadratic vote.
type QuadraticVoteRecord struct {
	ProposalID string    `json:"proposal_id"`
	Voter      Address   `json:"voter"`
	Tokens     uint64    `json:"tokens"`
	Approve    bool      `json:"approve"`
	Timestamp  time.Time `json:"timestamp"`
}

var qvMu sync.Mutex

// QuadraticWeight converts a token amount into quadratic voting power.
func QuadraticWeight(tokens uint64) uint64 {
	return uint64(math.Sqrt(float64(tokens)))
}

// SubmitQuadraticVote records a vote weighted by the square root of staked tokens.
// It checks the caller balance via the ledger and stores the vote in the global KV store.
func SubmitQuadraticVote(pID string, voter Address, tokens uint64, approve bool) error {
	led := CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	if led.BalanceOf(voter) < tokens {
		return fmt.Errorf("insufficient balance")
	}
	weight := QuadraticWeight(tokens)
	key := fmt.Sprintf("qvote:%s:%s", pID, hex.EncodeToString(voter[:]))
	rec := QuadraticVoteRecord{ProposalID: pID, Voter: voter, Tokens: weight, Approve: approve, Timestamp: time.Now().UTC()}
	raw, _ := json.Marshal(rec)
	qvMu.Lock()
	defer qvMu.Unlock()
	return CurrentStore().Set([]byte(key), raw)
}

// QuadraticResults tallies the quadratic votes for a proposal.
func QuadraticResults(pID string) (forWeight, againstWeight uint64, err error) {
	prefix := []byte("qvote:" + pID + ":")
	it := CurrentStore().Iterator(prefix, nil)
	defer it.Close()
	for it.Next() {
		var rec QuadraticVoteRecord
		if err = json.Unmarshal(it.Value(), &rec); err != nil {
			return
		}
		if rec.Approve {
			forWeight += rec.Tokens
		} else {
			againstWeight += rec.Tokens
		}
	}
	err = it.Error()
	return
}
