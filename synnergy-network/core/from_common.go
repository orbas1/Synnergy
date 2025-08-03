//go:build !tokens
// +build !tokens

package core

import "github.com/ethereum/go-ethereum/common"

// FromCommon converts an Ethereum common.Address to the Synnergy Address type.
func FromCommon(a common.Address) Address {
	var out Address
	copy(out[:], a.Bytes())
	return out
}
