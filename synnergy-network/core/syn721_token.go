package core

import "sync"

// SYN721Metadata holds metadata for a single NFT token
// URI references external data, Data stores JSON or other string metadata,
// History keeps previous metadata versions.
type SYN721Metadata struct {
	URI     string
	Data    string
	History []string
}

// SYN721Token implements a minimal NFT token standard similar to ERC-721.
type SYN721Token struct {
	BaseToken
	owners    map[uint64]Address
	approvals map[uint64]Address
	metaStore map[uint64]SYN721Metadata
	nextID    uint64
	mu        sync.RWMutex
}

// NewSYN721Token creates a new instance with zero supply.
func NewSYN721Token(meta Metadata) *SYN721Token {
	bt := BaseToken{id: deriveID(meta.Standard), meta: meta, balances: NewBalanceTable()}
	return &SYN721Token{
		BaseToken: bt,
		owners:    make(map[uint64]Address),
		approvals: make(map[uint64]Address),
		metaStore: make(map[uint64]SYN721Metadata),
	}
}

// ID returns the token ID
func (t *SYN721Token) ID() TokenID { return t.BaseToken.id }

// Meta returns static metadata about the token standard
func (t *SYN721Token) Meta() Metadata { return t.BaseToken.meta }

// BalanceOf returns how many NFTs an address owns
func (t *SYN721Token) BalanceOf(a Address) uint64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var cnt uint64
	for _, owner := range t.owners {
		if owner == a {
			cnt++
		}
	}
	return cnt
}

// Allowance is always zero for SYN721 (approvals are per token)
func (t *SYN721Token) Allowance(owner, spender Address) uint64 { return 0 }

// Approve grants permission to transfer a specific NFT
func (t *SYN721Token) Approve(owner, spender Address, nftID uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.owners[nftID] != owner {
		return ErrInvalidAsset
	}
	t.approvals[nftID] = spender
	if t.BaseToken.ledger != nil {
		t.BaseToken.ledger.EmitApproval(t.BaseToken.id, owner, spender, 1)
	}
	return nil
}

// Transfer moves ownership of a specific NFT
func (t *SYN721Token) Transfer(from, to Address, nftID uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if owner := t.owners[nftID]; owner != from && t.approvals[nftID] != from {
		return ErrInvalidAsset
	}
	t.owners[nftID] = to
	delete(t.approvals, nftID)
	if t.BaseToken.ledger != nil {
		fee := t.BaseToken.gas.Calculate("OpTokenTransfer", 1)
		t.BaseToken.ledger.DeductGas(from, fee)
		t.BaseToken.ledger.EmitTransfer(t.BaseToken.id, from, to, 1)
	}
	return nil
}

// MintWithMeta mints a new NFT with metadata and returns the NFT ID
func (t *SYN721Token) MintWithMeta(to Address, md SYN721Metadata) (uint64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	id := t.nextID
	t.nextID++
	t.owners[id] = to
	t.metaStore[id] = md
	t.meta.TotalSupply++
	if t.BaseToken.ledger != nil {
		t.BaseToken.ledger.EmitTransfer(t.BaseToken.id, Address{}, to, 1)
	}
	return id, nil
}

// Mint satisfies the Token interface by minting an empty metadata NFT
func (t *SYN721Token) Mint(to Address, amount uint64) error {
	for i := uint64(0); i < amount; i++ {
		if _, err := t.MintWithMeta(to, SYN721Metadata{}); err != nil {
			return err
		}
	}
	return nil
}

// Burn removes an NFT from circulation
func (t *SYN721Token) Burn(from Address, nftID uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.owners[nftID] != from {
		return ErrInvalidAsset
	}
	delete(t.owners, nftID)
	delete(t.metaStore, nftID)
	delete(t.approvals, nftID)
	t.meta.TotalSupply--
	if t.BaseToken.ledger != nil {
		t.BaseToken.ledger.EmitTransfer(t.BaseToken.id, from, Address{}, 1)
	}
	return nil
}

// Metadata returns metadata for an NFT
func (t *SYN721Token) MetadataOf(id uint64) (SYN721Metadata, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	m, ok := t.metaStore[id]
	return m, ok
}

// UpdateMetadata replaces the metadata for an NFT
func (t *SYN721Token) UpdateMetadata(id uint64, md SYN721Metadata) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.metaStore[id]; !ok {
		return ErrInvalidAsset
	}
	prev := t.metaStore[id]
	if prev.Data != "" {
		md.History = append(prev.History, prev.Data)
	}
	t.metaStore[id] = md
	return nil
}
