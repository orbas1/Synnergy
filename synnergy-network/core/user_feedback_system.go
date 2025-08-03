package core

// user_feedback_system.go -- user feedback collection and reward engine
// ----------------------------------------------------------------------
// This module stores user feedback on-chain and exposes helper functions
// for submitting entries, fetching them by ID and listing all stored
// feedback. Entries are stored in the ledger state under the prefix
// "feedback:". New submissions are broadcast on the network so that
// replicators and indexers can react immediately.
//
// Rewarding positive contributions is optional; the FeedbackEngine can
// mint SYNN coins directly using the ledger MintToken helper. Ratings are
// 1–5 and callers are responsible for additional validation.
// ----------------------------------------------------------------------

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

//---------------------------------------------------------------------
// Data structures
//---------------------------------------------------------------------

// FeedbackEntry represents a single user submitted feedback record.
type FeedbackEntry struct {
	ID        string  `json:"id"`
	User      Address `json:"user"`
	Rating    uint8   `json:"rating"`
	Message   string  `json:"message"`
	Timestamp int64   `json:"ts"`
}

//---------------------------------------------------------------------
// Engine singleton
//---------------------------------------------------------------------

var feedbackOnce sync.Once
var feedbackEng *FeedbackEngine

// FeedbackEngine stores feedback using the provided ledger backend.
type FeedbackEngine struct {
	led StateRW
	mu  sync.Mutex
}

// InitFeedback sets up the global feedback engine with the given ledger.
func InitFeedback(led StateRW) { feedbackOnce.Do(func() { feedbackEng = &FeedbackEngine{led: led} }) }

// Feedback returns the initialised engine instance. It panics if InitFeedback
// has not been called.
func Feedback() *FeedbackEngine {
	if feedbackEng == nil {
		panic("feedback engine not initialised")
	}
	return feedbackEng
}

//---------------------------------------------------------------------
// Core operations
//---------------------------------------------------------------------

// Submit records a new feedback entry from a user and broadcasts the event.
// It returns the feedback ID which is a hex encoded SHA‑256 hash.
func (f *FeedbackEngine) Submit(user Address, rating uint8, msg string) (string, error) {
	if rating == 0 || rating > 5 {
		return "", errors.New("rating must be between 1 and 5")
	}
	if len(msg) == 0 {
		return "", errors.New("message required")
	}
	b := make([]byte, len(user)+8+len(msg))
	copy(b, user[:])
	binary.LittleEndian.PutUint64(b[len(user):], uint64(time.Now().UnixNano()))
	copy(b[len(user)+8:], []byte(msg))
	sum := sha256.Sum256(b)
	id := hex.EncodeToString(sum[:])

	entry := FeedbackEntry{ID: id, User: user, Rating: rating, Message: msg, Timestamp: time.Now().Unix()}
	raw, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}

	key := append([]byte("feedback:"), sum[:]...)
	if err := f.led.SetState(key, raw); err != nil {
		return "", err
	}
	if err := Broadcast("feedback", raw); err != nil {
		return "", err
	}
	return id, nil
}

// Get retrieves a single feedback entry by ID.
func (f *FeedbackEngine) Get(id string) (FeedbackEntry, error) {
	var out FeedbackEntry
	b, err := hex.DecodeString(id)
	if err != nil {
		return out, fmt.Errorf("bad id: %w", err)
	}
	key := append([]byte("feedback:"), b...)
	raw, err := f.led.GetState(key)
	if err != nil {
		return out, err
	}
	if len(raw) == 0 {
		return out, errors.New("not found")
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, err
	}
	return out, nil
}

// List returns all stored feedback entries.
func (f *FeedbackEngine) List() ([]FeedbackEntry, error) {
	it := f.led.PrefixIterator([]byte("feedback:"))
	var out []FeedbackEntry
	for it.Next() {
		var e FeedbackEntry
		if err := json.Unmarshal(it.Value(), &e); err == nil {
			out = append(out, e)
		}
	}
	return out, it.Error()
}

// Reward grants SYNN tokens to the user who submitted the given feedback ID.
func (f *FeedbackEngine) Reward(id string, amt uint64) error {
	if amt == 0 {
		return errors.New("amount must be >0")
	}
	entry, err := f.Get(id)
	if err != nil {
		return err
	}
	return f.led.Mint(entry.User, amt)
}

//---------------------------------------------------------------------
// END user_feedback_system.go
//---------------------------------------------------------------------
