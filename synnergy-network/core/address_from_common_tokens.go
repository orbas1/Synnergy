//go:build tokens
// +build tokens

package core

import "github.com/ethereum/go-ethereum/common"

// FromCommon converts a go-ethereum common.Address into the core Address type.
// It copies the 20-byte address into our Address array.
func FromCommon(a common.Address) Address {
	var out Address
	copy(out[:], a.Bytes())
	return out
}
