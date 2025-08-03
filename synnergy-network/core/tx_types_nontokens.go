//go:build !tokens

package core

// TxType identifiers for basic transaction categories.
// These values are provided here for builds that do not
// include the full token implementation (build tag "tokens").
// When the "tokens" tag is enabled, an extended set of
// transaction logic is compiled in transactions.go.

const (
	// TxPayment represents a standard currency transfer.
	TxPayment TxType = iota + 1
	// TxContractCall identifies a generic smart contract invocation.
	TxContractCall
	// TxReversal denotes an authority-approved transaction reversal.
	TxReversal
)
