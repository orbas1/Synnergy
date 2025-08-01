package core

import "sync"

// SYN700Token implements the Intellectual Property token standard.
type SYN700Token struct {
	*BaseToken
	mu sync.RWMutex
}

// NewSYN700Token constructs an empty SYN700 token instance.
func NewSYN700Token(meta Metadata) *SYN700Token {
	return &SYN700Token{
		BaseToken: &BaseToken{meta: meta, balances: NewBalanceTable()},
	}
}

func (t *SYN700Token) RegisterIPAsset(id string, meta IPMetadata, owner Address) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, err := RegisterIPAsset(id, meta, owner)
	return err
}

func (t *SYN700Token) TransferIPOwnership(id string, from, to Address, share uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return TransferIPOwnership(id, from, to, share)
}

func (t *SYN700Token) CreateLicense(id string, lic *License) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return CreateLicense(id, lic)
}

func (t *SYN700Token) RevokeLicense(id string, licensee Address) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return RevokeLicense(id, licensee)
}

func (t *SYN700Token) RecordRoyalty(id string, licensee Address, amount uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return RecordRoyalty(id, licensee, amount)
}
