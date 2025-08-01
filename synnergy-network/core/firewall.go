package core

import (
	"errors"
	"net"
	"sync"
)

// firewall.go - simple address/token/IP firewall for Synnergy Network

var (
	ErrAddrBlocked  = errors.New("address blocked by firewall")
	ErrTokenBlocked = errors.New("token blocked by firewall")
	ErrIPBlocked    = errors.New("ip blocked by firewall")
)

// Firewall maintains runtime block lists used by consensus and the ledger.
// It is concurrency safe and integrates with transaction validation.
type Firewall struct {
	mu        sync.RWMutex
	addresses map[Address]struct{}
	tokens    map[TokenID]struct{}
	ips       map[string]struct{}
}

// NewFirewall constructs an empty firewall instance.
func NewFirewall() *Firewall {
	return &Firewall{
		addresses: make(map[Address]struct{}),
		tokens:    make(map[TokenID]struct{}),
		ips:       make(map[string]struct{}),
	}
}

// BlockAddress prevents an account from sending or receiving funds.
func (fw *Firewall) BlockAddress(a Address) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.addresses[a] = struct{}{}
}

// UnblockAddress removes an address from the block list.
func (fw *Firewall) UnblockAddress(a Address) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	delete(fw.addresses, a)
}

// IsAddressBlocked checks if an address is blocked.
func (fw *Firewall) IsAddressBlocked(a Address) bool {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	_, ok := fw.addresses[a]
	return ok
}

// BlockToken disallows transfers of a specific token.
func (fw *Firewall) BlockToken(id TokenID) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.tokens[id] = struct{}{}
}

// UnblockToken removes a token from the block list.
func (fw *Firewall) UnblockToken(id TokenID) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	delete(fw.tokens, id)
}

// IsTokenBlocked checks if a token is blocked.
func (fw *Firewall) IsTokenBlocked(id TokenID) bool {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	_, ok := fw.tokens[id]
	return ok
}

// BlockIP bans a peer IP address from network participation.
func (fw *Firewall) BlockIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return errors.New("invalid ip")
	}
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.ips[ip] = struct{}{}
	return nil
}

// UnblockIP removes an IP from the banned list.
func (fw *Firewall) UnblockIP(ip string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	delete(fw.ips, ip)
}

// IsIPBlocked checks if an IP is blocked.
func (fw *Firewall) IsIPBlocked(ip string) bool {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	_, ok := fw.ips[ip]
	return ok
}

// FirewallRules snapshots all current rules for inspection.
type FirewallRules struct {
	Addresses []Address
	Tokens    []TokenID
	IPs       []string
}

// ListRules returns the blocked addresses, tokens and IPs.
func (fw *Firewall) ListRules() FirewallRules {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	rules := FirewallRules{}
	for a := range fw.addresses {
		rules.Addresses = append(rules.Addresses, a)
	}
	for t := range fw.tokens {
		rules.Tokens = append(rules.Tokens, t)
	}
	for ip := range fw.ips {
		rules.IPs = append(rules.IPs, ip)
	}
	return rules
}

// CheckTx verifies whether a transaction violates any firewall rule.
func (fw *Firewall) CheckTx(tx *Transaction) error {
	if fw == nil || tx == nil {
		return nil
	}
	if fw.IsAddressBlocked(tx.From) || fw.IsAddressBlocked(tx.To) {
		return ErrAddrBlocked
	}
	for _, tt := range tx.TokenTransfers {
		if fw.IsAddressBlocked(tt.From) || fw.IsAddressBlocked(tt.To) {
			return ErrAddrBlocked
		}
		if fw.IsTokenBlocked(tt.Token) {
			return ErrTokenBlocked
		}
	}
	return nil
}
