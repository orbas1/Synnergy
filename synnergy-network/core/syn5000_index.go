package core

import "time"

// BetRecord defines bet metadata for the SYN5000 standard.
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

// GamblingToken exposes methods of the SYN5000 token.
type GamblingToken interface {
	Token
	PlaceBet(addr Address, game string, amount uint64, odds float64) (uint64, error)
	ResolveBet(id uint64, won bool) (uint64, error)
	BetInfo(id uint64) (BetRecord, bool)
}

var _ GamblingToken = (*SYN5000Token)(nil)
