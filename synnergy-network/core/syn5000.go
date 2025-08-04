package core

import (
	"fmt"
	"sync"
	"time"
)

// BetRecord stores betting activity for SYN5000 tokens.
type BetRecord struct {
	ID       uint64
	GameType string
	Bettor   Address
	Amount   uint64
	Odds     float64
	Placed   time.Time
	Resolved bool
	Won      bool
	Payout   uint64
}

// SYN5000Token represents the gambling token standard.
type SYN5000Token struct {
	*BaseToken
	ledger *Ledger
	gas    GasCalculator

	mu     sync.RWMutex
	bets   map[uint64]*BetRecord
	nextID uint64
}

// NewSYN5000Token creates the token and registers it with the registry.
func NewSYN5000Token(meta Metadata, init map[Address]uint64, ledger *Ledger, gas GasCalculator) *SYN5000Token {
	meta.Standard = StdSYN5000
	tok, _ := (Factory{}).Create(meta, init)
	bt := tok.(*BaseToken)
	bt.ledger = ledger
	bt.gas = gas
	st := &SYN5000Token{BaseToken: bt, ledger: ledger, gas: gas, bets: make(map[uint64]*BetRecord)}
	RegisterToken(st)
	return st
}

// PlaceBet transfers amount to escrow and records the bet.
func (t *SYN5000Token) PlaceBet(addr Address, game string, amount uint64, odds float64) (uint64, error) {
	if odds <= 0 {
		return 0, fmt.Errorf("invalid odds")
	}
	if err := t.Transfer(addr, AddressZero, amount); err != nil {
		return 0, err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	id := t.nextID
	t.nextID++
	t.bets[id] = &BetRecord{ID: id, GameType: game, Bettor: addr, Amount: amount, Odds: odds, Placed: time.Now().UTC()}
	return id, nil
}

// ResolveBet finalises the bet outcome and mints winnings if applicable.
func (t *SYN5000Token) ResolveBet(id uint64, won bool) (uint64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	bet, ok := t.bets[id]
	if !ok {
		return 0, fmt.Errorf("bet not found")
	}
	if bet.Resolved {
		return bet.Payout, nil
	}
	bet.Resolved = true
	bet.Won = won
	if won {
		payout := uint64(float64(bet.Amount) * bet.Odds)
		bet.Payout = payout
		_ = t.Mint(bet.Bettor, payout)
	}
	return bet.Payout, nil
}

// BetInfo returns information about a specific bet.
func (t *SYN5000Token) BetInfo(id uint64) (BetRecord, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	b, ok := t.bets[id]
	if !ok {
		return BetRecord{}, false
	}
	cp := *b
	return cp, true
}
