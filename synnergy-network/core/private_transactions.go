package core

// Private transactions helpers provide lightweight encryption and
// submission logic so that payloads can be hidden from the public
// ledger. Functions here integrate with the existing TxPool and
// security package.

import (
	"encoding/hex"
	"errors"
)

// EncryptTxPayload encrypts tx.Payload using the supplied key.
// The encrypted data is stored in tx.EncryptedPayload and the
// original payload is cleared. The tx.Private flag is set.
func EncryptTxPayload(tx *Transaction, key []byte) error {
	if tx == nil {
		return errors.New("nil transaction")
	}
	if len(tx.Payload) == 0 {
		return errors.New("empty payload")
	}
	blob, err := Encrypt(key, tx.Payload, nil)
	if err != nil {
		return err
	}
	tx.EncryptedPayload = blob
	tx.Payload = nil
	tx.Private = true
	return nil
}

// DecryptTxPayload decrypts tx.EncryptedPayload using the key and
// returns the plaintext payload. It does not modify the transaction
// so that callers may decide whether to reveal the data.
func DecryptTxPayload(tx *Transaction, key []byte) ([]byte, error) {
	if tx == nil {
		return nil, errors.New("nil transaction")
	}
	if !tx.Private {
		return nil, errors.New("transaction not private")
	}
	if len(tx.EncryptedPayload) == 0 {
		return nil, errors.New("missing encrypted payload")
	}
	return Decrypt(key, tx.EncryptedPayload, nil)
}

// SubmitPrivateTx validates and inserts the transaction into the
// provided pool. It assumes the payload has already been encrypted
// and the transaction signed.
func SubmitPrivateTx(pool *TxPool, tx *Transaction) error {
	if pool == nil || tx == nil {
		return errors.New("pool or tx is nil")
	}
	return pool.AddTx(tx)
}

// EncodeEncryptedHex is a helper that returns the encrypted payload
// as a hex string for easy transport or storage.
func EncodeEncryptedHex(tx *Transaction) (string, error) {
	if tx == nil {
		return "", errors.New("nil transaction")
	}
	if len(tx.EncryptedPayload) == 0 {
		return "", errors.New("no encrypted payload")
	}
	return hex.EncodeToString(tx.EncryptedPayload), nil
}
