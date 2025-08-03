//go:build tokens
// +build tokens

package core

// TxType categorizes high-level transaction kinds used across the Synnergy core.
//
// The full token build exposes all transaction variants, so this file is only
// compiled when the `tokens` build tag is supplied. Lightweight builds rely on
// `tx_types_nontokens.go`, which defines the minimal subset required by the
// wallet manager and other non-token modules.
type TxType uint8

const (

	// TxPayment transfers value between addresses.
	TxPayment TxType = iota + 1
	// TxContractCall executes a smart contract.
	TxContractCall
	// TxReversal denotes an authority-approved reversal of a previous
	// transaction. The recipient refunds the sender minus a protocol fee.
	TxReversal


)
