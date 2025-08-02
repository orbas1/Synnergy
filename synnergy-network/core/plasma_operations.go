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

// PlasmaBridgeDeposit records a token deposit into the plasma bridge.
type PlasmaBridgeDeposit struct {
	Nonce  uint64      `json:"nonce"`
	From   Address     `json:"from"`
	Token  TokenID     `json:"token"`
	Amount uint64      `json:"amount"`
	Block  PlasmaBlock `json:"block"`
	Time   int64       `json:"time"`
}

// PlasmaBridgeExit tracks an exit request from the plasma chain.
type PlasmaBridgeExit struct {
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

// BridgeCoordinator manages deposits and exits.
type BridgeCoordinator struct {
	Ledger StateRW
	Net    PlasmaBroadcaster
	mu     sync.Mutex
	nonce  uint64
}

var (
	plasmaOpOnce sync.Once
	bridgeCoord  *BridgeCoordinator
)

// InitPlasmaCoordinator sets up the global plasma coordinator.
func InitPlasmaCoordinator(led StateRW, net PlasmaBroadcaster) {
	plasmaOpOnce.Do(func() { bridgeCoord = &BridgeCoordinator{Ledger: led, Net: net} })
}

// Plasma returns the globally configured coordinator.
func PlasmaCoordinatorInstance() *BridgeCoordinator { return bridgeCoord }

// Deposit moves tokens from the user account to the plasma bridge.
func (p *BridgeCoordinator) Deposit(from Address, token TokenID, amount uint64, blk PlasmaBlock) (PlasmaBridgeDeposit, error) {
	if p == nil {
		return PlasmaBridgeDeposit{}, errors.New("plasma not initialised")
	}
	if amount == 0 {
		return PlasmaBridgeDeposit{}, errors.New("zero amount")
	}
	tok, ok := GetToken(token)
	if !ok {
		return PlasmaBridgeDeposit{}, errors.New("token unknown")
	}
	bridge := plasmaBridgeAccount(token)
	if err := tok.Transfer(from, bridge, amount); err != nil {
		return PlasmaBridgeDeposit{}, err
	}

	p.mu.Lock()
	p.nonce++
	nonce := p.nonce
	p.mu.Unlock()

	dep := PlasmaBridgeDeposit{Nonce: nonce, From: from, Token: token, Amount: amount, Block: blk, Time: time.Now().Unix()}
	key := depositKeyPlasma(nonce)
	p.Ledger.SetState(key, plasmaJSON(dep))
	if p.Net != nil {
		_ = p.Net.Broadcast("plasma_deposit", dep)
	}
	return dep, nil
}

// StartExit records an exit intent which must later be finalised.
func (p *BridgeCoordinator) StartExit(owner Address, token TokenID, amount uint64, blk PlasmaBlock) (PlasmaBridgeExit, error) {
	if p == nil {
		return PlasmaBridgeExit{}, errors.New("plasma not initialised")
	}
	if amount == 0 {
		return PlasmaBridgeExit{}, errors.New("zero amount")
	}
	bridge := plasmaBridgeAccount(token)
	bal := p.Ledger.BalanceOf(bridge)
	if bal < amount {
		return PlasmaBridgeExit{}, fmt.Errorf("insufficient bridge balance: %d", bal)
	}

	p.mu.Lock()
	p.nonce++
	nonce := p.nonce
	p.mu.Unlock()

	ex := PlasmaBridgeExit{Nonce: nonce, Owner: owner, Token: token, Amount: amount, Block: blk}
	key := exitKeyPlasma(nonce)
	p.Ledger.SetState(key, plasmaJSON(ex))
	if p.Net != nil {
		_ = p.Net.Broadcast("plasma_exit", ex)
	}
	return ex, nil
}

// FinalizeExit releases tokens from the bridge to the owner.
func (p *BridgeCoordinator) FinalizeExit(nonce uint64) error {
	if p == nil {
		return errors.New("plasma not initialised")
	}
	raw, err := p.Ledger.GetState(exitKeyPlasma(nonce))
	if err != nil || raw == nil {
		return errors.New("exit not found")
	}
	var ex PlasmaBridgeExit
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
	bridge := plasmaBridgeAccount(ex.Token)
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
func (p *BridgeCoordinator) GetExit(nonce uint64) (PlasmaBridgeExit, error) {
	if p == nil {
		return PlasmaBridgeExit{}, errors.New("plasma not initialised")
	}
	raw, err := p.Ledger.GetState(exitKeyPlasma(nonce))
	if err != nil || raw == nil {
		return PlasmaBridgeExit{}, errors.New("exit not found")
	}
	var ex PlasmaBridgeExit
	if err := json.Unmarshal(raw, &ex); err != nil {
		return PlasmaBridgeExit{}, err
	}
	return ex, nil
}

// ListExits returns all exits for the given owner.
func (p *BridgeCoordinator) ListExits(owner Address) ([]PlasmaBridgeExit, error) {
	if p == nil {
		return nil, errors.New("plasma not initialised")
	}
	it := p.Ledger.PrefixIterator([]byte("plasma:exit:"))
	var out []PlasmaBridgeExit
	for it.Next() {
		var ex PlasmaBridgeExit
		if err := json.Unmarshal(it.Value(), &ex); err != nil {
			continue
		}
		if ex.Owner == owner {
			out = append(out, ex)
		}
	}
	return out, nil
}


// plasmaBridgeAccount returns the address used as a bridge for Plasma
// operations. It includes a distinct prefix to avoid collisions with other
// bridge types.
func plasmaBridgeAccount(token TokenID) Address {
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
