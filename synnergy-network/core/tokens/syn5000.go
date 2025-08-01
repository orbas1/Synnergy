package core

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Bet represents a single gambling bet placed using a SYN5000 token.
type Bet struct {
	ID       string    `json:"id"`
	GameType string    `json:"game_type"`
	Amount   uint64    `json:"amount"`
	Player   Address   `json:"player"`
	Winner   Address   `json:"winner"`
	Outcome  string    `json:"outcome"`
	Settled  bool      `json:"settled"`
	Created  time.Time `json:"created"`
}

// SYN5000Token extends BaseToken with gambling utilities.
type SYN5000Token struct {
	*BaseToken
}

var (
	synLedger *Ledger
	betMu     sync.RWMutex
	bets      = make(map[string]*Bet)
)

// InitSYN5000 wires a ledger instance for bet escrow and state storage.
func InitSYN5000(led *Ledger) {
	if led != nil {
		synLedger = led
	}
}

// NewSYN5000 constructs a new SYN5000 token around an existing BaseToken.
func NewSYN5000(bt *BaseToken) *SYN5000Token {
	return &SYN5000Token{BaseToken: bt}
}

// PlaceBet deducts the stake from the player and records the bet on the ledger.
func (t *SYN5000Token) PlaceBet(player Address, gameType string, amount uint64) (Bet, error) {
	if synLedger == nil {
		return Bet{}, fmt.Errorf("ledger not initialised")
	}
	if err := t.Transfer(player, AddressZero, amount); err != nil {
		return Bet{}, err
	}
	b := &Bet{ID: uuid.New().String(), GameType: gameType, Amount: amount, Player: player, Created: time.Now().UTC()}
	enc, _ := json.Marshal(b)
	if err := synLedger.SetState([]byte("syn5000:"+b.ID), enc); err != nil {
		return Bet{}, err
	}
	betMu.Lock()
	bets[b.ID] = b
	betMu.Unlock()
	return *b, nil
}

// ResolveBet finalises a bet and pays the winnings to the winner address.
func (t *SYN5000Token) ResolveBet(id, outcome string, winner Address) error {
	if synLedger == nil {
		return fmt.Errorf("ledger not initialised")
	}
	betMu.Lock()
	b, ok := bets[id]
	betMu.Unlock()
	if !ok {
		raw, err := synLedger.GetState([]byte("syn5000:" + id))
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, &b); err != nil {
			return err
		}
	}
	if b.Settled {
		return fmt.Errorf("bet already settled")
	}
	b.Outcome = outcome
	b.Winner = winner
	b.Settled = true
	if err := synLedger.Transfer(AddressZero, winner, b.Amount); err != nil {
		return err
	}
	enc, _ := json.Marshal(b)
	if err := synLedger.SetState([]byte("syn5000:"+b.ID), enc); err != nil {
		return err
	}
	betMu.Lock()
	bets[b.ID] = b
	betMu.Unlock()
	return nil
}

// GetBet retrieves a bet by ID.
func (t *SYN5000Token) GetBet(id string) (Bet, error) {
	betMu.RLock()
	b, ok := bets[id]
	betMu.RUnlock()
	if ok {
		return *b, nil
	}
	if synLedger == nil {
		return Bet{}, fmt.Errorf("ledger not initialised")
	}
	raw, err := synLedger.GetState([]byte("syn5000:" + id))
	if err != nil {
		return Bet{}, err
	}
	if err := json.Unmarshal(raw, &b); err != nil {
		return Bet{}, err
	}
	betMu.Lock()
	bets[id] = b
	betMu.Unlock()
	return *b, nil
}

// ListBets returns all known bets.
func (t *SYN5000Token) ListBets() ([]Bet, error) {
	if synLedger == nil {
		return nil, fmt.Errorf("ledger not initialised")
	}
	betMu.RLock()
	out := make([]Bet, 0, len(bets))
	for _, b := range bets {
		out = append(out, *b)
	}
	betMu.RUnlock()
	it := synLedger.PrefixIterator([]byte("syn5000:"))
	for it.Next() {
		id := string(it.Key())[8:]
		if _, ok := bets[id]; !ok {
			var b Bet
			if err := json.Unmarshal(it.Value(), &b); err != nil {
				return nil, err
			}
			betMu.Lock()
			bets[id] = &b
			betMu.Unlock()
			out = append(out, b)
		}
	}
	return out, it.Error()
}
