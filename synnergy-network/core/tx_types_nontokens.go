//go:build !tokens

package core

// TxType identifiers for basic transaction categories.
//
// In builds that do not include the full token stack (the "tokens" build tag
// is absent) the comprehensive transaction types defined in transactions.go are
// not compiled.  This file provides the minimal subset of types required by
// lightweight components such as wallet management so that code depending on a
// basic payment transfer still compiles.

// TxType enumerates high level transaction categories.
// It mirrors the definition in transactions.go when the "tokens" tag is set.
type TxType uint8

const (
	// TxPayment represents a standard currency transfer.
	TxPayment TxType = iota + 1
	// TxContractCall identifies a generic smart contract invocation.
	TxContractCall
	// TxReversal denotes an authority-approved transaction reversal.
	TxReversal
)
