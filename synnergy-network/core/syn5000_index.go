package core

// GamblingToken exposes methods of the SYN5000 token.
type GamblingToken interface {
	Token
	PlaceBet(addr Address, game string, amount uint64, odds float64) (uint64, error)
	ResolveBet(id uint64, won bool) (uint64, error)
	BetInfo(id uint64) (BetRecord, bool)
}

var _ GamblingToken = (*SYN5000Token)(nil)
