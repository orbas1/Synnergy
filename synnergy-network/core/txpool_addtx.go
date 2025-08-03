package core

import "errors"

// AddTx inserts a transaction into the pool with basic validation and
// duplicate checks. It initialises internal maps and slices if necessary to
// avoid nil dereferences.
func (tp *TxPool) AddTx(tx *Transaction) error {
	if tp == nil {
		return errors.New("txpool not initialised")
	}
	if tx == nil {
		return errors.New("nil transaction")
	}
	tp.mu.Lock()
	defer tp.mu.Unlock()
	if tp.lookup == nil {
		tp.lookup = make(map[Hash]*Transaction)
	}
	if tp.queue == nil {
		tp.queue = make([]*Transaction, 0)
	}
	if _, exists := tp.lookup[tx.Hash]; exists {
		return errors.New("tx already in pool")
	}
	tp.lookup[tx.Hash] = tx
	tp.queue = append(tp.queue, tx)
	return nil
}
