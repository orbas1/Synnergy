package core

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ReputationEvent records an adjustment to a user's reputation score.
type ReputationEvent struct {
	Timestamp   time.Time `json:"ts"`
	Delta       int64     `json:"delta"`
	Description string    `json:"desc"`
}

// ReputationRecord tracks score, trust level and history for an address.
type ReputationRecord struct {
	Score  int64             `json:"score"`
	Level  string            `json:"level"`
	Events []ReputationEvent `json:"events"`
}

// ReputationEngine manages SYN1500 reputation scores backed by the SYN-REP token.
type ReputationEngine struct {
	ledger *Ledger
	mu     sync.Mutex
	data   map[Address]*ReputationRecord
}

var (
	repEngine *ReputationEngine
	repOnce   sync.Once
	// syn1500ReputationTokenID derives from the SYN1500 standard constant.
	syn1500ReputationTokenID = deriveID(StdSYN1500)
)

// InitReputationEngine creates the singleton reputation manager.
func InitReputationEngine(led *Ledger) {
	repOnce.Do(func() {
		repEngine = &ReputationEngine{ledger: led, data: make(map[Address]*ReputationRecord)}
	})
}

// Reputation returns the initialised engine instance.
func Reputation() *ReputationEngine { return repEngine }

func (e *ReputationEngine) record(addr Address) *ReputationRecord {
	r, ok := e.data[addr]
	if !ok {
		r = &ReputationRecord{Level: "Bronze"}
		e.data[addr] = r
	}
	return r
}

// AddActivity adjusts reputation based on activity description.
func (e *ReputationEngine) AddActivity(addr Address, delta int64, desc string) error {
	if e == nil {
		return fmt.Errorf("reputation engine not initialised")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	rec := e.record(addr)
	rec.Score += delta
	rec.Events = append(rec.Events, ReputationEvent{Timestamp: time.Now().UTC(), Delta: delta, Description: desc})
	e.updateLevel(rec)
	tok, ok := TokenLedger[syn1500ReputationTokenID]
	if ok {
		if delta > 0 {
			_ = tok.Mint(addr, uint64(delta))
		} else if delta < 0 {
			_ = tok.Burn(addr, uint64(-delta))
		}
	}
	return nil
}

// Endorse adds a positive endorsement from another user.
func (e *ReputationEngine) Endorse(addr, from Address, points int64, comment string) error {
	note := fmt.Sprintf("endorsement from %s: %s", from.String(), comment)
	return e.AddActivity(addr, points, note)
}

// Penalize deducts reputation due to negative behaviour.
func (e *ReputationEngine) Penalize(addr Address, points int64, reason string) error {
	if points < 0 {
		points = -points
	}
	note := fmt.Sprintf("penalty: %s", reason)
	return e.AddActivity(addr, -points, note)
}

// Score returns the current reputation score.
func (e *ReputationEngine) Score(addr Address) int64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.record(addr).Score
}

// Level returns the trust level for the address.
func (e *ReputationEngine) Level(addr Address) string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.record(addr).Level
}

// History returns a copy of the reputation event log.
func (e *ReputationEngine) History(addr Address) []ReputationEvent {
	e.mu.Lock()
	defer e.mu.Unlock()
	rec := e.record(addr)
	out := make([]ReputationEvent, len(rec.Events))
	copy(out, rec.Events)
	return out
}

func (e *ReputationEngine) updateLevel(r *ReputationRecord) {
	switch {
	case r.Score >= 1000:
		r.Level = "Platinum"
	case r.Score >= 500:
		r.Level = "Gold"
	case r.Score >= 250:
		r.Level = "Silver"
	default:
		r.Level = "Bronze"
	}
}

// --- VM opcode helpers -------------------------------------------------------

// RepAddActivity is exposed to the VM for reputation scoring.
func RepAddActivity(ctx *Context, addr Address, delta int64, desc string) error {
	return Reputation().AddActivity(addr, delta, desc)
}

// RepEndorse exposes Endorse for opcodes.
func RepEndorse(ctx *Context, addr, from Address, pts int64, comment string) error {
	return Reputation().Endorse(addr, from, pts, comment)
}

// RepPenalize exposes Penalize for opcodes.
func RepPenalize(ctx *Context, addr Address, pts int64, reason string) error {
	return Reputation().Penalize(addr, pts, reason)
}

// RepScore pushes the score onto the stack.
func RepScore(ctx *Context, addr Address) int64 { return Reputation().Score(addr) }

// RepLevel returns the trust level string.
func RepLevel(ctx *Context, addr Address) string { return Reputation().Level(addr) }

// RepHistory returns reputation events as JSON bytes.
func RepHistory(ctx *Context, addr Address) []byte {
	events := Reputation().History(addr)
	b, _ := json.Marshal(events)
	return b
}
