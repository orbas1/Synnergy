package core

import (
	"fmt"
)

// WalletManager wraps Ledger and HDWallet helpers to perform high level wallet operations.
type WalletManager struct {
	ledger *Ledger
}

// NewWalletManager creates a manager bound to the given ledger.
func NewWalletManager(l *Ledger) *WalletManager {
	return &WalletManager{ledger: l}
}

// Create generates a random HD wallet with the given entropy bits and returns it along
// with the mnemonic phrase. The wallet is not persisted to disk.
func (wm *WalletManager) Create(bits int) (*HDWallet, string, error) {
	return NewRandomWallet(bits)
}

// Import constructs a wallet from the provided mnemonic and optional passphrase.
func (wm *WalletManager) Import(mnemonic, passphrase string) (*HDWallet, error) {
	return WalletFromMnemonic(mnemonic, passphrase)
}

// Balance returns the SYNN balance for the given address using the manager ledger.
func (wm *WalletManager) Balance(addr Address) uint64 {
	if wm.ledger == nil {
		return 0
	}
	return wm.ledger.BalanceOf(addr)
}

// Transfer signs and submits a payment transaction from the wallet to the target address.
// It also updates the ledger immediately by calling Ledger.Transfer.
func (wm *WalletManager) Transfer(w *HDWallet, account, index uint32, to Address, amount, gasPrice uint64) (*Transaction, error) {
	if wm.ledger == nil {
		return nil, fmt.Errorf("ledger not initialised")
	}
	from, err := w.NewAddress(account, index)
	if err != nil {
		return nil, err
	}
	tx := &Transaction{
		Type:     TxPayment,
		To:       to,
		Value:    amount,
		GasPrice: gasPrice,
		Nonce:    wm.ledger.NonceOf(from),
	}
	if err := w.SignTx(tx, account, index, gasPrice); err != nil {
		return nil, err
	}
	if err := wm.ledger.Transfer(from, to, amount); err != nil {
		return nil, err
	}
	wm.ledger.AddToPool(tx)
	return tx, nil
}
