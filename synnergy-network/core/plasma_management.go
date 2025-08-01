package core

// plasma_management.go - minimal plasma chain coordinator integrated with ledger and consensus.
//
// Responsibilities:
//  - track user deposits into a plasma contract account
//  - allow exits back to the main chain
//  - store submitted plasma blocks for fraud proofs
//
// Dependencies: ledger, tokens and optional consensus broadcast.

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// PlasmaMetaBlock holds a minimal commitment for a plasma child-chain block.
type PlasmaMetaBlock struct {
	Height    uint64
	Root      []byte // merkle root of transactions
	Timestamp int64
}

// PlasmaDepositRecord records a deposit locked into the plasma contract.
type PlasmaDepositRecord struct {
	ID        uint64
	From      Address
	Token     TokenID
	Amount    uint64
	Timestamp int64
}

// PlasmaManager is a singleton managing deposits and submitted blocks.
type PlasmaManager struct {
	Ledger    StateRW
	Consensus interface {
		Broadcast(topic string, data interface{}) error
	}

	mu          sync.Mutex
	nextDeposit uint64
	blocks      []PlasmaMetaBlock
}

var (
	plasmaMgrOnce sync.Once
	plasmaMgr     *PlasmaManager
)

// InitPlasmaManager initialises the plasma manager with a ledger and optional consensus engine.
func InitPlasmaManager(led StateRW, cons interface {
	Broadcast(string, interface{}) error
}) {
	plasmaMgrOnce.Do(func() { plasmaMgr = &PlasmaManager{Ledger: led, Consensus: cons} })
}

// Plasma returns the active plasma manager instance.
func PlasmaManagerInstance() *PlasmaManager { return plasmaMgr }

// Deposit locks tokens into the plasma contract and records the deposit.
func (pm *PlasmaManager) Deposit(from Address, token TokenID, amount uint64) (PlasmaDepositRecord, error) {
	if amount == 0 {
		return PlasmaDepositRecord{}, errors.New("zero amount")
	}
	tok, ok := GetToken(token)
	if !ok {
		return PlasmaDepositRecord{}, errors.New("token unknown")
	}
	if err := tok.Transfer(from, plasmaAccount(token), amount); err != nil {
		return PlasmaDepositRecord{}, err
	}
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.nextDeposit++
	dep := PlasmaDepositRecord{ID: pm.nextDeposit, From: from, Token: token, Amount: amount, Timestamp: time.Now().Unix()}
	pm.Ledger.SetState(depKey(dep.ID), mustJSON(dep))
	return dep, nil
}

// Withdraw releases a previously recorded deposit back to the user.
func (pm *PlasmaManager) Withdraw(id uint64, to Address) error {
	raw, _ := pm.Ledger.GetState(depKey(id))
	if len(raw) == 0 {
		return errors.New("deposit not found")
	}
	if wd, _ := pm.Ledger.GetState(wdKey(id)); len(wd) != 0 {
		return errors.New("already withdrawn")
	}
	var dep PlasmaDepositRecord
	_ = json.Unmarshal(raw, &dep)
	tok, ok := GetToken(dep.Token)
	if !ok {
		return errors.New("token unknown")
	}
	if err := tok.Transfer(plasmaAccount(dep.Token), to, dep.Amount); err != nil {
		return err
	}
	pm.Ledger.SetState(wdKey(id), []byte{1})
	return nil
}

// SubmitBlock stores a plasma block commitment and broadcasts it via consensus.
func (pm *PlasmaManager) SubmitBlock(root []byte) (PlasmaMetaBlock, error) {
	if len(root) == 0 {
		return PlasmaMetaBlock{}, errors.New("empty root")
	}
	pm.mu.Lock()
	defer pm.mu.Unlock()
	blk := PlasmaMetaBlock{Height: uint64(len(pm.blocks) + 1), Root: root, Timestamp: time.Now().Unix()}
	pm.blocks = append(pm.blocks, blk)
	pm.Ledger.SetState(blockKey(blk.Height), mustJSON(blk))
	if pm.Consensus != nil {
		_ = pm.Consensus.Broadcast("plasma_block", blk)
	}
	return blk, nil
}

// GetBlock fetches a previously submitted plasma block.
func (pm *PlasmaManager) GetBlock(h uint64) (PlasmaMetaBlock, error) {
	raw, _ := pm.Ledger.GetState(blockKey(h))
	if len(raw) == 0 {
		return PlasmaMetaBlock{}, errors.New("unknown block")
	}
	var blk PlasmaMetaBlock
	_ = json.Unmarshal(raw, &blk)
	return blk, nil
}

// --- helper keys and accounts ---

func plasmaAccount(t TokenID) Address {
	var a Address
	copy(a[:4], []byte("PLAS"))
	binary.BigEndian.PutUint32(a[4:], uint32(t))
	return a
}

func depKey(id uint64) []byte  { return append([]byte("plasma:dep:"), uint64ToBytes(id)...) }
func wdKey(id uint64) []byte   { return append([]byte("plasma:wd:"), uint64ToBytes(id)...) }
func blockKey(h uint64) []byte { return append([]byte("plasma:blk:"), uint64ToBytes(h)...) }

//---------------------------------------------------------------------
// END plasma_management.go
//---------------------------------------------------------------------
