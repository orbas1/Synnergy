package core

import (
	"crypto/sha256"
	"encoding/binary"
	"sync"
)

// LightningChannelID uniquely identifies a payment channel.
type LightningChannelID [32]byte

// LightningChannel represents a two party balance state.
type LightningChannel struct {
	ID       LightningChannelID
	PartyA   Address
	PartyB   Address
	Token    TokenID
	BalanceA uint64
	BalanceB uint64
	Nonce    uint64
}

// LightningNode enables off-chain micro payments across multiple channels.
type LightningNode struct {
	ledger   *Ledger
	mu       sync.RWMutex
	channels map[LightningChannelID]*LightningChannel
	nonce    uint64
}

var (
	lnOnce sync.Once
	ln     *LightningNode
)

// InitLightning initialises the lightning engine with the provided ledger.
func InitLightning(ledger *Ledger) {
	lnOnce.Do(func() {
		ln = &LightningNode{ledger: ledger, channels: make(map[LightningChannelID]*LightningChannel)}
	})
}

// Lightning returns the global node instance.
func Lightning() *LightningNode { return ln }

// OpenChannel escrows funds and records a new channel.
func (l *LightningNode) OpenChannel(a, b Address, token TokenID, amtA, amtB uint64) (LightningChannelID, error) {
	tok, ok := GetToken(token)
	if !ok {
		return LightningChannelID{}, ErrInvalidAsset
	}
	l.mu.Lock()
	l.nonce++
	nonce := l.nonce
	l.mu.Unlock()
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, nonce)
	h := sha256.Sum256(append(append(a.Bytes(), b.Bytes()...), buf...))
	var id LightningChannelID
	copy(id[:], h[:])
	esc := channelEscrow(id)
	if amtA > 0 {
		if err := tok.Transfer(a, esc, amtA); err != nil {
			return id, err
		}
	}
	if amtB > 0 {
		if err := tok.Transfer(b, esc, amtB); err != nil {
			return id, err
		}
	}
	l.mu.Lock()
	l.channels[id] = &LightningChannel{ID: id, PartyA: a, PartyB: b, Token: token, BalanceA: amtA, BalanceB: amtB}
	l.mu.Unlock()
	return id, nil
}

// RoutePayment updates channel balances off-chain.
func (l *LightningNode) RoutePayment(id LightningChannelID, from Address, amount uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	ch, ok := l.channels[id]
	if !ok {
		return ErrInvalidAsset
	}
	if from == ch.PartyA {
		if ch.BalanceA < amount {
			return ErrInvalidAsset
		}
		ch.BalanceA -= amount
		ch.BalanceB += amount
	} else if from == ch.PartyB {
		if ch.BalanceB < amount {
			return ErrInvalidAsset
		}
		ch.BalanceB -= amount
		ch.BalanceA += amount
	} else {
		return ErrInvalidAsset
	}
	ch.Nonce++
	return nil
}

// CloseChannel releases escrowed funds according to latest balances.
func (l *LightningNode) CloseChannel(id LightningChannelID) error {
	l.mu.Lock()
	ch, ok := l.channels[id]
	if !ok {
		l.mu.Unlock()
		return ErrInvalidAsset
	}
	delete(l.channels, id)
	l.mu.Unlock()
	tok, ok := GetToken(ch.Token)
	if !ok {
		return ErrInvalidAsset
	}
	esc := channelEscrow(id)
	if ch.BalanceA > 0 {
		if err := tok.Transfer(esc, ch.PartyA, ch.BalanceA); err != nil {
			return err
		}
	}
	if ch.BalanceB > 0 {
		if err := tok.Transfer(esc, ch.PartyB, ch.BalanceB); err != nil {
			return err
		}
	}
	return nil
}

// ListChannels returns active channels snapshot.
func (l *LightningNode) ListChannels() []LightningChannel {
	l.mu.RLock()
	out := make([]LightningChannel, 0, len(l.channels))
	for _, ch := range l.channels {
		out = append(out, *ch)
	}
	l.mu.RUnlock()
	return out
}

func channelEscrow(id LightningChannelID) Address {
	var a Address
	copy(a[:], id[:20])
	return a
}
