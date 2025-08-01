package smartcontracts

import (
	core "synnergy-network/core"
)

// WalletContract demonstrates using opcodes from the dispatcher.
type WalletContract struct{}

const (
	opNewRandomWallet    = core.Opcode(0x1D0001)
	opWalletFromMnemonic = core.Opcode(0x1D0002)
	opSignTx             = core.Opcode(0x1D0006)
)

// Bytecodes exposes hex representations of key wallet opcodes.
func Bytecodes() map[string]string {
	return map[string]string{
		"NewRandomWallet":    opNewRandomWallet.Hex(),
		"WalletFromMnemonic": opWalletFromMnemonic.Hex(),
		"SignTx":             opSignTx.Hex(),
	}
}
