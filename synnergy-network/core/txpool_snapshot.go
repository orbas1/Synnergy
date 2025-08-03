//go:build !tokens

package core

// Snapshot returns a copy of all pending transactions in the pool.
// It acquires a read lock to allow concurrent access while taking the snapshot.
func (tp *TxPool) Snapshot() []*Transaction {
	if tp == nil {
		return nil
	}

	tp.mu.RLock()
	defer tp.mu.RUnlock()

	if len(tp.queue) == 0 {
		return nil
	}

	// Make a copy to avoid exposing internal slice allowing callers to mutate state.
	list := make([]*Transaction, len(tp.queue))
	copy(list, tp.queue)
	return list
}
