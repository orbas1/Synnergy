package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// PlasmaBlock represents a block reference on the plasma chain.
type PlasmaBlock struct {
	Height    uint64   `json:"height"`
	Root      [32]byte `json:"root"`
	Timestamp int64    `json:"timestamp"`
}

// PlasmaDeposit records a token deposit into the plasma bridge.
type PlasmaDeposit struct {
	Nonce  uint64      `json:"nonce"`
	From   Address     `json:"from"`
	Token  TokenID     `json:"token"`
	Amount uint64      `json:"amount"`
	Block  PlasmaBlock `json:"block"`
	Time   int64       `json:"time"`
}

// PlasmaExit tracks an exit request from the plasma chain.
type PlasmaExit struct {
	Nonce     uint64      `json:"nonce"`
	Owner     Address     `json:"owner"`
	Token     TokenID     `json:"token"`
	Amount    uint64      `json:"amount"`
	Block     PlasmaBlock `json:"block"`
	Finalized bool        `json:"finalized"`
}

// Broadcaster is a minimal interface for consensus/network integration.
type PlasmaBroadcaster interface {
	Broadcast(topic string, msg interface{}) error
}

// PlasmaCoordinator manages deposits and exits.
type PlasmaCoordinator struct {
	Ledger StateRW
	Net    PlasmaBroadcaster
	mu     sync.Mutex
	nonce  uint64
}

var (
	plasmaOnce sync.Once
	plasma     *PlasmaCoordinator
)

// InitPlasma sets up the global plasma coordinator.
func InitPlasma(led StateRW, net PlasmaBroadcaster) {
	plasmaOnce.Do(func() { plasma = &PlasmaCoordinator{Ledger: led, Net: net} })
}

// Plasma returns the globally configured coordinator.
func Plasma() *PlasmaCoordinator { return plasma }

// Deposit moves tokens from the user account to the plasma bridge.
func (p *PlasmaCoordinator) Deposit(from Address, token TokenID, amount uint64, blk PlasmaBlock) (PlasmaDeposit, error) {
	if p == nil {
		return PlasmaDeposit{}, errors.New("plasma not initialised")
	}
	if amount == 0 {
		return PlasmaDeposit{}, errors.New("zero amount")
	}
	tok, ok := GetToken(token)
	if !ok {
		return PlasmaDeposit{}, errors.New("token unknown")
	}
	bridge := plasmaAccount(token)
	if err := tok.Transfer(from, bridge, amount); err != nil {
		return PlasmaDeposit{}, err
	}

	p.mu.Lock()
	p.nonce++
	nonce := p.nonce
	p.mu.Unlock()

	dep := PlasmaDeposit{Nonce: nonce, From: from, Token: token, Amount: amount, Block: blk, Time: time.Now().Unix()}
	key := depositKeyPlasma(nonce)
	p.Ledger.SetState(key, plasmaJSON(dep))
	if p.Net != nil {
		_ = p.Net.Broadcast("plasma_deposit", dep)
	}
	return dep, nil
}

// StartExit records an exit intent which must later be finalised.
func (p *PlasmaCoordinator) StartExit(owner Address, token TokenID, amount uint64, blk PlasmaBlock) (PlasmaExit, error) {
	if p == nil {
		return PlasmaExit{}, errors.New("plasma not initialised")
	}
	if amount == 0 {
		return PlasmaExit{}, errors.New("zero amount")
	}
	bridge := plasmaAccount(token)
	bal := p.Ledger.BalanceOf(bridge)
	if bal < amount {
		return PlasmaExit{}, fmt.Errorf("insufficient bridge balance: %d", bal)
	}

	p.mu.Lock()
	p.nonce++
	nonce := p.nonce
	p.mu.Unlock()

	ex := PlasmaExit{Nonce: nonce, Owner: owner, Token: token, Amount: amount, Block: blk}
	key := exitKeyPlasma(nonce)
	p.Ledger.SetState(key, plasmaJSON(ex))
	if p.Net != nil {
		_ = p.Net.Broadcast("plasma_exit", ex)
	}
	return ex, nil
}

// FinalizeExit releases tokens from the bridge to the owner.
func (p *PlasmaCoordinator) FinalizeExit(nonce uint64) error {
	if p == nil {
		return errors.New("plasma not initialised")
	}
	raw, err := p.Ledger.GetState(exitKeyPlasma(nonce))
	if err != nil || raw == nil {
		return errors.New("exit not found")
	}
	var ex PlasmaExit
	if err := json.Unmarshal(raw, &ex); err != nil {
		return err
	}
	if ex.Finalized {
		return errors.New("already finalised")
	}
	tok, ok := GetToken(ex.Token)
	if !ok {
		return errors.New("token unknown")
	}
	bridge := plasmaAccount(ex.Token)
	if err := tok.Transfer(bridge, ex.Owner, ex.Amount); err != nil {
		return err
	}
	ex.Finalized = true
	p.Ledger.SetState(exitKeyPlasma(nonce), plasmaJSON(ex))
	if p.Net != nil {
		_ = p.Net.Broadcast("plasma_finalized", ex)
	}
	return nil
}

// GetExit fetches a previously recorded exit.
func (p *PlasmaCoordinator) GetExit(nonce uint64) (PlasmaExit, error) {
	if p == nil {
		return PlasmaExit{}, errors.New("plasma not initialised")
	}
	raw, err := p.Ledger.GetState(exitKeyPlasma(nonce))
	if err != nil || raw == nil {
		return PlasmaExit{}, errors.New("exit not found")
	}
	var ex PlasmaExit
	if err := json.Unmarshal(raw, &ex); err != nil {
		return PlasmaExit{}, err
	}
	return ex, nil
}

// ListExits returns all exits for the given owner.
func (p *PlasmaCoordinator) ListExits(owner Address) ([]PlasmaExit, error) {
	if p == nil {
		return nil, errors.New("plasma not initialised")
	}
	it := p.Ledger.PrefixIterator([]byte("plasma:exit:"))
	var out []PlasmaExit
	for it.Next() {
		var ex PlasmaExit
		if err := json.Unmarshal(it.Value(), &ex); err != nil {
			continue
		}
		if ex.Owner == owner {
			out = append(out, ex)
		}
	}
	return out, nil
}

func plasmaAccount(token TokenID) Address {
	var a Address
	copy(a[:4], []byte("PLSM"))
	a[4] = byte(token >> 24)
	a[5] = byte(token >> 16)
	a[6] = byte(token >> 8)
	a[7] = byte(token)
	return a
}

func depositKeyPlasma(n uint64) []byte { return append([]byte("plasma:dep:"), uint64ToBytes(n)...) }
func exitKeyPlasma(n uint64) []byte    { return append([]byte("plasma:exit:"), uint64ToBytes(n)...) }

func plasmaJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
