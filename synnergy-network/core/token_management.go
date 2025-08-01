package core

import (
	"sync"
)

// TokenManager provides high level helpers for creating and manipulating tokens
// through the ledger and VM. It acts as a bridge between the token registry and
// other subsystems such as consensus and transaction processing.

type TokenManager struct {
	ledger *Ledger
	gas    GasCalculator
	mu     sync.RWMutex
}

// NewTokenManager initialises a manager bound to the given ledger and gas model.
func NewTokenManager(l *Ledger, g GasCalculator) *TokenManager {
	return &TokenManager{ledger: l, gas: g}
}

// Create mints a new token and registers it with the ledger and registry.
func (tm *TokenManager) Create(meta Metadata, init map[Address]uint64) (TokenID, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tok, err := (Factory{}).Create(meta, init)
	if err != nil {
		return 0, err
	}
	bt := tok.(*BaseToken)
	bt.ledger = tm.ledger
	bt.gas = tm.gas
	if tm.ledger.tokens == nil {
		tm.ledger.tokens = make(map[TokenID]Token)
	}
	tm.ledger.tokens[bt.id] = bt
	return bt.id, nil
}

// Transfer moves balances between addresses for the given token.
func (tm *TokenManager) Transfer(id TokenID, from, to Address, amount uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	return tok.Transfer(from, to, amount)
}

// Mint creates new supply for the specified token.
func (tm *TokenManager) Mint(id TokenID, to Address, amount uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	return tok.Mint(to, amount)
}

// Burn removes supply from the specified holder.
func (tm *TokenManager) Burn(id TokenID, from Address, amount uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	return tok.Burn(from, amount)
}

// Approve sets an allowance for a spender.
func (tm *TokenManager) Approve(id TokenID, owner, spender Address, amount uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	return tok.Approve(owner, spender, amount)
}

// BalanceOf returns the balance of an address for a token.
func (tm *TokenManager) BalanceOf(id TokenID, addr Address) (uint64, error) {
	tok, ok := GetToken(id)
	if !ok {
		return 0, ErrInvalidAsset
	}
	return tok.BalanceOf(addr), nil
}

// Mint721 mints a new NFT with metadata and returns the NFT identifier.
func (tm *TokenManager) Mint721(id TokenID, to Address, meta SYN721Metadata) (uint64, error) {
	tok, ok := GetToken(id)
	if !ok {
		return 0, ErrInvalidAsset
	}
	nft, ok := tok.(*SYN721Token)
	if !ok {
		return 0, ErrInvalidAsset
	}
	return nft.MintWithMeta(to, meta)
}

// Transfer721 transfers ownership of a specific NFT token.
func (tm *TokenManager) Transfer721(id TokenID, from, to Address, nftID uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	nft, ok := tok.(*SYN721Token)
	if !ok {
		return ErrInvalidAsset
	}
	return nft.Transfer(from, to, nftID)
}

// Burn721 burns a specific NFT token.
func (tm *TokenManager) Burn721(id TokenID, owner Address, nftID uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	nft, ok := tok.(*SYN721Token)
	if !ok {
		return ErrInvalidAsset
	}
	return nft.Burn(owner, nftID)
}

// Metadata721 retrieves metadata for a given NFT.
func (tm *TokenManager) Metadata721(id TokenID, nftID uint64) (SYN721Metadata, error) {
	tok, ok := GetToken(id)
	if !ok {
		return SYN721Metadata{}, ErrInvalidAsset
	}
	nft, ok := tok.(*SYN721Token)
	if !ok {
		return SYN721Metadata{}, ErrInvalidAsset
	}
	m, _ := nft.MetadataOf(nftID)
	return m, nil
}

// UpdateMetadata721 updates metadata for a given NFT.
func (tm *TokenManager) UpdateMetadata721(id TokenID, nftID uint64, meta SYN721Metadata) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	nft, ok := tok.(*SYN721Token)
	if !ok {
		return ErrInvalidAsset
	}
	return nft.UpdateMetadata(nftID, meta)
}
