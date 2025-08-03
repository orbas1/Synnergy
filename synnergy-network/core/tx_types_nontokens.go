//go:build !tokens

package core

// TxType categorizes transaction kinds for non-token builds.
type TxType uint8

const (
	// TxPayment transfers value between addresses.
	TxPayment TxType = iota + 1
	// TxContractCall executes a smart contract.
	TxContractCall
	// TxReversal denotes an authority-approved reversal of a previous transaction.
	TxReversal
)
