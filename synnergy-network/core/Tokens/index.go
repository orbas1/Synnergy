package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// Address is a 20 byte array mirroring the core Address type.
type Address [20]byte

// AccessInfo defines access rights and reward state.
type AccessInfo struct {
	Tier         uint8
	MaxUsage     uint64
	UsageCount   uint64
	Expiry       int64
	RewardPoints uint64
}

// SYN500 exposes the extended functionality of the SYN500 utility token.
type SYN500 interface {
	TokenInterfaces
	GrantAccess(addr Address, tier uint8, max uint64, expiry int64)
	UpdateAccess(addr Address, tier uint8, max uint64, expiry int64)
	RevokeAccess(addr Address)
	RecordUsage(addr Address, points uint64) error
	RedeemReward(addr Address, points uint64) error
	RewardBalance(addr Address) uint64
	Usage(addr Address) uint64
	AccessInfoOf(addr Address) (AccessInfo, bool)
}
