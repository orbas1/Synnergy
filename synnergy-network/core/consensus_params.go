//go:build !tokens

package core

import "math/big"

var InitialReward *big.Int

const RewardHalvingPeriod = 200_000

func init() {
	var ok bool
	InitialReward, ok = new(big.Int).SetString("102400000000000000000", 10)
	if !ok {
		panic("invalid InitialReward value")
	}
}
