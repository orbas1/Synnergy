package core

// TxType categorizes high-level transaction kinds used across the Synnergy core.
type TxType uint8

const (
	// TxPayment transfers value between addresses.
	TxPayment TxType = iota + 1
	// TxContractCall executes a smart contract.
	TxContractCall
	// TxReversal denotes a reversal of a previous transaction.
	TxReversal
)
