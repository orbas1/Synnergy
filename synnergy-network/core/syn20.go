package core

import "sync"

// SYN20Token extends core.BaseToken with pause and freeze capabilities.
type SYN20Token struct {
	*BaseToken
	paused bool
	freeze map[Address]bool
	mu     sync.RWMutex
}

// NewSYN20Token creates a new SYN20 token instance using core.Factory.
func NewSYN20Token(meta Metadata, init map[Address]uint64) (*SYN20Token, error) {
	if meta.Standard == 0 {
		meta.Standard = StdSYN20
	}
	f := Factory{}
	tok, err := f.Create(meta, init)
	if err != nil {
		return nil, err
	}
	return &SYN20Token{
		BaseToken: tok.(*BaseToken),
		freeze:    make(map[Address]bool),
	}, nil
}

// Pause disables all token transfers.
func (t *SYN20Token) Pause() {
	t.mu.Lock()
	t.paused = true
	t.mu.Unlock()
}

// Unpause re-enables transfers.
func (t *SYN20Token) Unpause() {
	t.mu.Lock()
	t.paused = false
	t.mu.Unlock()
}

// IsPaused returns whether transfers are disabled.
func (t *SYN20Token) IsPaused() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.paused
}

// Freeze marks an account as frozen.
func (t *SYN20Token) Freeze(a Address) {
	t.mu.Lock()
	t.freeze[a] = true
	t.mu.Unlock()
}

// Thaw removes an account from the freeze list.
func (t *SYN20Token) Thaw(a Address) {
	t.mu.Lock()
	delete(t.freeze, a)
	t.mu.Unlock()
}

// Transfer overrides BaseToken.Transfer to enforce pause and freeze rules.
func (t *SYN20Token) Transfer(from, to Address, amt uint64) error {
	t.mu.RLock()
	paused := t.paused
	frozen := t.freeze[from] || t.freeze[to]
	t.mu.RUnlock()
	if paused || frozen {
		return ErrInvalidAsset
	}
	return t.BaseToken.Transfer(from, to, amt)
}

// TransferWithMemo performs a transfer and ignores the memo parameter.
func (t *SYN20Token) TransferWithMemo(from, to Address, amt uint64, memo string) error {
	return t.Transfer(from, to, amt)
}

// BulkTransfer sends tokens to multiple recipients.
func (t *SYN20Token) BulkTransfer(from Address, tos []Address, amts []uint64) error {
	if len(tos) != len(amts) {
		return ErrInvalidAsset
	}
	for i, to := range tos {
		if err := t.Transfer(from, to, amts[i]); err != nil {
			return err
		}
	}
	return nil
}

// BulkApprove grants allowances to multiple spenders.
func (t *SYN20Token) BulkApprove(owner Address, spenders []Address, amts []uint64) error {
	if len(spenders) != len(amts) {
		return ErrInvalidAsset
	}
	for i, s := range spenders {
		if err := t.Approve(owner, s, amts[i]); err != nil {
			return err
		}
	}
	return nil
}
