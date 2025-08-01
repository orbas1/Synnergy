package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// RewardTokenInterface defines the extended methods of the SYN600
// reward token standard without importing core types.
type RewardTokenInterface interface {
	TokenInterfaces
	Stake(addr any, amount uint64, duration int64) error
	Unstake(addr any) error
	AddEngagement(addr any, points uint64) error
	EngagementOf(addr any) uint64
	DistributeStakingRewards(rate uint64) error
}
