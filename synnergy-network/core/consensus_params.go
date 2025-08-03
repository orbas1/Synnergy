//go:build !tokens

package core

import "math/big"

// RewardHalvingPeriod defines how many main blocks occur before the block
// reward is halved. The value mirrors the setting used in consensus.go but is
// provided here without build tags so that packages depending on these
// constants can compile in all build configurations.
const RewardHalvingPeriod = 200_000 // blocks (main)

// InitialReward is the starting block reward prior to any halving events. It
// matches the value established in consensus.go and is expressed as a big.Int
// to preserve precision for token amounts.
var InitialReward *big.Int

func init() {
	var ok bool
	InitialReward, ok = new(big.Int).SetString("102400000000000000000", 10)
	if !ok {
		panic("invalid InitialReward value")
	}
}
