package core

// GamblingManager is a thin wrapper around the SYN5000 token implementation.
type GamblingManager struct {
	token *SYN5000Token
}

func NewGamblingManager(tok *SYN5000Token) *GamblingManager {
	return &GamblingManager{token: tok}
}

func (gm *GamblingManager) PlaceBet(player Address, gameType string, amt uint64) (Bet, error) {
	return gm.token.PlaceBet(player, gameType, amt)
}

func (gm *GamblingManager) ResolveBet(id, outcome string, winner Address) error {
	return gm.token.ResolveBet(id, outcome, winner)
}

func (gm *GamblingManager) GetBet(id string) (Bet, error) {
	return gm.token.GetBet(id)
}

func (gm *GamblingManager) ListBets() ([]Bet, error) {
	return gm.token.ListBets()
}
