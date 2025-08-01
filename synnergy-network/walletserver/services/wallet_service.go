package services

import (
	core "synnergy-network/core"
	"synnergy-network/walletserver/smartcontracts"

)

// WalletService wraps core wallet operations used by the HTTP API.
type WalletService struct{}

func NewService() *WalletService { return &WalletService{} }

func (ws *WalletService) CreateWallet(bits int) (*core.HDWallet, string, error) {
	return core.NewRandomWallet(bits)
}

func (ws *WalletService) ImportWallet(mnemonic, passphrase string) (*core.HDWallet, error) {
	return core.WalletFromMnemonic(mnemonic, passphrase)
}

func (ws *WalletService) DeriveAddress(w *core.HDWallet, account, index uint32) (core.Address, error) {
	return w.NewAddress(account, index)
}

func (ws *WalletService) SignTransaction(w *core.HDWallet, tx *core.Transaction, account, index uint32, gas uint64) error {
	return w.SignTx(tx, account, index, gas)
}

// Opcodes returns the hex mapping of wallet opcode helpers.
func (ws *WalletService) Opcodes() map[string]string {
	return smartcontracts.Bytecodes()
}

