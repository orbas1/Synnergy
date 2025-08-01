package core

import (
	"fmt"
	"sync"
)

// SYN223Token implements the SYN223 safe-transfer token standard.
// It embeds BaseToken and adds whitelist, blacklist and multi-signature support
// to ensure transfers only go to compatible contracts and authorised addresses.
type SYN223Token struct {
	*BaseToken
	mu           sync.RWMutex
	whitelist    map[Address]bool
	blacklist    map[Address]bool
	requiredSigs int
}

// NewSYN223Token creates a SYN223 token with the supplied metadata and initial balances.
func NewSYN223Token(meta Metadata, init map[Address]uint64) *SYN223Token {
	bt := &BaseToken{id: deriveID(meta.Standard), meta: meta, balances: NewBalanceTable()}
	tok := &SYN223Token{
		BaseToken:    bt,
		whitelist:    make(map[Address]bool),
		blacklist:    make(map[Address]bool),
		requiredSigs: 2,
	}
	for a, v := range init {
		tok.balances.Set(tok.id, a, v)
		tok.meta.TotalSupply += v
	}
	RegisterToken(tok)
	return tok
}

// AddToWhitelist allows an address to receive tokens.
func (t *SYN223Token) AddToWhitelist(a Address) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.whitelist[a] = true
}

// RemoveFromWhitelist removes an address from the whitelist.
func (t *SYN223Token) RemoveFromWhitelist(a Address) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.whitelist, a)
}

// AddToBlacklist prevents an address from receiving tokens.
func (t *SYN223Token) AddToBlacklist(a Address) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.blacklist[a] = true
}

// RemoveFromBlacklist removes an address from the blacklist.
func (t *SYN223Token) RemoveFromBlacklist(a Address) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.blacklist, a)
}

// IsWhitelisted checks if an address is on the whitelist.
func (t *SYN223Token) IsWhitelisted(a Address) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.whitelist[a]
}

// IsBlacklisted checks if an address is blacklisted.
func (t *SYN223Token) IsBlacklisted(a Address) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.blacklist[a]
}

// SetRequiredSignatures sets how many signatures are required for transfers.
func (t *SYN223Token) SetRequiredSignatures(n int) {
	if n <= 0 {
		n = 1
	}
	t.mu.Lock()
	t.requiredSigs = n
	t.mu.Unlock()
}

// SafeTransfer performs a transfer after verifying whitelist/blacklist and
// contract compatibility. It also enforces a basic multi-signature requirement
// based solely on the number of supplied signatures.
func (t *SYN223Token) SafeTransfer(from, to Address, amount uint64, sigs ...[]byte) error {
	if len(sigs) < t.requiredSigs {
		return fmt.Errorf("insufficient signatures: require %d", t.requiredSigs)
	}
	if t.IsBlacklisted(to) {
		return fmt.Errorf("recipient blacklisted")
	}
	if len(t.whitelist) > 0 && !t.IsWhitelisted(to) {
		return fmt.Errorf("recipient not whitelisted")
	}
	if t.isContract(to) && !t.supportsTokenReceiver(to) {
		return fmt.Errorf("recipient contract cannot receive SYN223 tokens")
	}
	return t.BaseToken.Transfer(from, to, amount)
}

// isContract checks if the address belongs to a deployed contract.
func (t *SYN223Token) isContract(a Address) bool {
	reg := GetContractRegistry()
	if reg == nil {
		return false
	}
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	_, ok := reg.byAddr[a]
	return ok
}

// supportsTokenReceiver returns true if the contract advertises SYN223 support.
// In this prototype it simply checks ledger state key "syn223:rcv:<addr>".
func (t *SYN223Token) supportsTokenReceiver(a Address) bool {
	if t.ledger == nil {
		return false
	}
	key := append([]byte("syn223:rcv:"), a.Bytes()...)
	b, err := t.ledger.GetState(key)
	return err == nil && len(b) > 0 && b[0] == 1
}
