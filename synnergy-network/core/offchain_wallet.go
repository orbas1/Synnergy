package core

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// OffChainWallet wraps HDWallet for offline signing and storage utilities.
type OffChainWallet struct {
	*HDWallet
	logger *log.Logger
}

// NewOffChainWallet creates a fresh wallet and returns the mnemonic.
func NewOffChainWallet(entropyBits int, lg *log.Logger) (*OffChainWallet, string, error) {
	w, mnemonic, err := NewRandomWallet(entropyBits)
	if err != nil {
		return nil, "", err
	}
	if lg == nil {
		lg = log.Default()
	}
	return &OffChainWallet{HDWallet: w, logger: lg}, mnemonic, nil
}

// OffChainWalletFromMnemonic imports an existing mnemonic.
func OffChainWalletFromMnemonic(mnemonic, passphrase string, lg *log.Logger) (*OffChainWallet, error) {
	w, err := WalletFromMnemonic(mnemonic, passphrase)
	if err != nil {
		return nil, err
	}
	if lg == nil {
		lg = log.Default()
	}
	return &OffChainWallet{HDWallet: w, logger: lg}, nil
}

// SignOffline signs the transaction without broadcasting it.
func (ow *OffChainWallet) SignOffline(tx *Transaction, account, index uint32, gasPrice uint64) error {
	if ow == nil || ow.HDWallet == nil {
		return fmt.Errorf("nil off-chain wallet")
	}
	return ow.SignTx(tx, account, index, gasPrice)
}

// StoreSignedTx writes the signed transaction to path in JSON form.
func StoreSignedTx(tx *Transaction, path string) error {
	if tx == nil {
		return fmt.Errorf("nil transaction")
	}
	out, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o600)
}

// LoadSignedTx reads a transaction JSON file.
func LoadSignedTx(path string) (*Transaction, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tx Transaction
	if err := json.Unmarshal(raw, &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}

// BroadcastSignedTx sends the signed transaction to the current ledger pool.
func BroadcastSignedTx(tx *Transaction) error {
	if tx == nil {
		return fmt.Errorf("nil transaction")
	}
	l := CurrentLedger()
	if l == nil {
		return fmt.Errorf("ledger not initialised")
	}
	l.AddToPool(tx)
	return nil
}
