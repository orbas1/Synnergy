package smartcontracts

import (
	core "synnergy-network/core"
)

// WalletContract demonstrates using opcodes from the dispatcher.
type WalletContract struct{}

const (
	opNewRandomWallet     = core.Opcode(0x1D0001)
	opWalletFromMnemonic  = core.Opcode(0x1D0002)
	opNewHDWalletFromSeed = core.Opcode(0x1D0003)
	opPrivateKey          = core.Opcode(0x1D0004)
	opNewAddress          = core.Opcode(0x1D0005)
	opSignTx              = core.Opcode(0x1D0006)
)

// Bytecodes exposes hex representations of key wallet opcodes.
// Bytecodes returns a map of wallet-related opcodes to their hex values.
func Bytecodes() map[string]string {
	return map[string]string{
		"NewRandomWallet":     opNewRandomWallet.Hex(),
		"WalletFromMnemonic":  opWalletFromMnemonic.Hex(),
		"NewHDWalletFromSeed": opNewHDWalletFromSeed.Hex(),
		"PrivateKey":          opPrivateKey.Hex(),
		"NewAddress":          opNewAddress.Hex(),
		"SignTx":              opSignTx.Hex(),

	}
}
