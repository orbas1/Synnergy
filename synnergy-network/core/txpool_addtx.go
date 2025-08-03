//go:build ignore

package core

import "fmt"

// AddTx inserts a transaction into the pool with minimal validation.
// It initialises internal maps if necessary to avoid nil dereferences.
func (tp *TxPool) AddTx(tx *Transaction) error {
	if tp == nil || tx == nil {
		return fmt.Errorf("txpool or tx nil")
	}
	tp.mu.Lock()
	defer tp.mu.Unlock()
	if tp.lookup == nil {
		tp.lookup = make(map[Hash]*Transaction)
	}
	tp.lookup[tx.Hash] = tx
	tp.queue = append(tp.queue, tx)
	return nil
}
