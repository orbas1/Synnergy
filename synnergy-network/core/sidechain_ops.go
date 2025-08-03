package core

// sidechain_ops.go -- management helpers for sidechain lifecycle

import "errors"

// PauseSidechain flags a sidechain as paused. Transactions and headers
// will be rejected until ResumeSidechain is called.
func (sc *SidechainCoordinator) PauseSidechain(id SidechainID) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	meta, err := sc.getMeta(id)
	if err != nil {
		return err
	}
	if meta.Paused {
		return errors.New("already paused")
	}
	meta.Paused = true
	return sc.Ledger.SetState(metaKey(id), mustJSON(meta))
}

// ResumeSidechain re-activates a previously paused sidechain.
func (sc *SidechainCoordinator) ResumeSidechain(id SidechainID) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	meta, err := sc.getMeta(id)
	if err != nil {
		return err
	}
	if !meta.Paused {
		return errors.New("not paused")
	}
	meta.Paused = false
	return sc.Ledger.SetState(metaKey(id), mustJSON(meta))
}

// UpdateSidechainValidators changes the validator set and threshold for an
// existing sidechain.
func (sc *SidechainCoordinator) UpdateSidechainValidators(id SidechainID, threshold uint8, validators [][]byte) error {
	if threshold == 0 || threshold > 100 {
		return errors.New("invalid threshold")
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	meta, err := sc.getMeta(id)
	if err != nil {
		return err
	}
	meta.Threshold = threshold
	meta.Validators = validators
	return sc.Ledger.SetState(metaKey(id), mustJSON(meta))
}

// RemoveSidechain deletes all metadata and pending deposits/headers for the
// given sidechain. This operation is irreversible.
func (sc *SidechainCoordinator) RemoveSidechain(id SidechainID) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if _, err := sc.getMeta(id); err != nil {
		return err
	}
	if err := sc.Ledger.DeleteState(metaKey(id)); err != nil {
		return err
	}
	hdrPrefix := append([]byte("sc:hdr:"), uint32ToBytes(uint32(id))...)
	hdrIt := sc.Ledger.PrefixIterator(hdrPrefix)
	defer hdrIt.Close()
	for hdrIt.Next() {
		if err := sc.Ledger.DeleteState(hdrIt.Key()); err != nil {
			return err
		}
	}
	if err := hdrIt.Error(); err != nil {
		return err
	}
	depPrefix := append([]byte("sc:dep:"), uint32ToBytes(uint32(id))...)
	depIt := sc.Ledger.PrefixIterator(depPrefix)
	defer depIt.Close()
	for depIt.Next() {
		if err := sc.Ledger.DeleteState(depIt.Key()); err != nil {
			return err
		}
	}
	return depIt.Error()
}
