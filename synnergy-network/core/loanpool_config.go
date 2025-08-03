package core

import "time"

// LoanPoolConfig defines configuration parameters for LoanPool.
type LoanPoolConfig struct {
	ElectorateSize       int                       `yaml:"electorate_size"`
	VotePeriod           time.Duration             `yaml:"vote_period"`
	SpamFee              uint64                    `yaml:"spam_fee"`
	RedistributeInterval time.Duration             `yaml:"redistribute_interval"`
	RedistributePerc     int                       `yaml:"redistribute_perc"`
	Rules                map[ProposalType]VoteRule `yaml:"rules"`
}
