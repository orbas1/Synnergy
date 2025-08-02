package core

import (
	"crypto/sha256"
	"encoding/json"
)

// HashTx computes and caches the transaction hash.
func (tx *Transaction) HashTx() Hash {
	b, _ := json.Marshal(tx)
	h := sha256.Sum256(b)
	tx.Hash = h
	return h
}
