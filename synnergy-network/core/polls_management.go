package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Poll represents a simple community poll stored in the global KV store.
type Poll struct {
	ID       string          `json:"id"`
	Question string          `json:"question"`
	Options  []string        `json:"options"`
	Counts   []uint64        `json:"counts"`
	Voters   map[string]bool `json:"voters"`
	Creator  Address         `json:"creator"`
	Deadline time.Time       `json:"deadline"`
	Closed   bool            `json:"closed"`
}

const pollPrefix = "poll:" // key prefix in the KV store

// CreatePoll registers a new poll. Duration defines how long voting is open.
func CreatePoll(question string, options []string, creator Address, duration time.Duration) (Poll, error) {
	if question == "" || len(options) < 2 {
		return Poll{}, fmt.Errorf("invalid poll parameters")
	}
	if CurrentStore() == nil {
		return Poll{}, fmt.Errorf("kv store not initialised")
	}

	p := Poll{
		ID:       uuid.New().String(),
		Question: question,
		Options:  append([]string(nil), options...),
		Counts:   make([]uint64, len(options)),
		Voters:   make(map[string]bool),
		Creator:  creator,
		Deadline: time.Now().Add(duration),
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return Poll{}, err
	}
	if err := CurrentStore().Set([]byte(pollPrefix+p.ID), raw); err != nil {
		return Poll{}, err
	}
	return p, nil
}

// VotePoll casts a vote on the given poll option index.
func VotePoll(id string, voter Address, option int) error {
	if CurrentStore() == nil {
		return fmt.Errorf("kv store not initialised")
	}
	raw, err := CurrentStore().Get([]byte(pollPrefix + id))
	if err != nil {
		return ErrNotFound
	}
	var p Poll
	if err := json.Unmarshal(raw, &p); err != nil {
		return err
	}
	if p.Closed || time.Now().After(p.Deadline) {
		return fmt.Errorf("poll closed")
	}
	addr := voter.Hex()
	if p.Voters[addr] {
		return fmt.Errorf("already voted")
	}
	if option < 0 || option >= len(p.Options) {
		return fmt.Errorf("invalid option")
	}
	if led := CurrentLedger(); led != nil {
		if led.BalanceOf(voter) == 0 {
			return ErrUnauthorized
		}
	}
	p.Voters[addr] = true
	p.Counts[option]++
	updated, _ := json.Marshal(&p)
	if err := CurrentStore().Set([]byte(pollPrefix+id), updated); err != nil {
		return err
	}
	return nil
}

// ClosePoll marks a poll as closed regardless of deadline.
func ClosePoll(id string) error {
	if CurrentStore() == nil {
		return fmt.Errorf("kv store not initialised")
	}
	raw, err := CurrentStore().Get([]byte(pollPrefix + id))
	if err != nil {
		return ErrNotFound
	}
	var p Poll
	if err := json.Unmarshal(raw, &p); err != nil {
		return err
	}
	if p.Closed {
		return ErrInvalidState
	}
	p.Closed = true
	updated, _ := json.Marshal(&p)
	return CurrentStore().Set([]byte(pollPrefix+id), updated)
}

// GetPoll retrieves a poll by ID.
func GetPoll(id string) (Poll, error) {
	if CurrentStore() == nil {
		return Poll{}, fmt.Errorf("kv store not initialised")
	}
	raw, err := CurrentStore().Get([]byte(pollPrefix + id))
	if err != nil {
		return Poll{}, ErrNotFound
	}
	var p Poll
	if err := json.Unmarshal(raw, &p); err != nil {
		return Poll{}, err
	}
	return p, nil
}

// ListPolls returns all polls in the store.
func ListPolls() ([]Poll, error) {
	if CurrentStore() == nil {
		return nil, fmt.Errorf("kv store not initialised")
	}
	it := CurrentStore().Iterator([]byte(pollPrefix), nil)
	var out []Poll
	for it.Next() {
		var p Poll
		if err := json.Unmarshal(it.Value(), &p); err == nil {
			out = append(out, p)
		}
	}
	if err := it.Error(); err != nil {
		return nil, err
	}
	return out, it.Close()
}
