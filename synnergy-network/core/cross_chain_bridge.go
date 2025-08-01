package core

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
)

// BridgeTransfer records a cross-chain transfer locked on this chain.
type BridgeTransfer struct {
	ID        string    `json:"id"`
	BridgeID  string    `json:"bridge_id"`
	From      Address   `json:"from"`
	To        Address   `json:"to"`
	Asset     AssetRef  `json:"asset"`
	Amount    uint64    `json:"amount"`
	Time      time.Time `json:"time"`
	Completed bool      `json:"completed"`
}

// StartBridgeTransfer locks assets from the caller and records a transfer.
func StartBridgeTransfer(ctx *Context, bridgeID string, asset AssetRef, to Address, amount uint64) (BridgeTransfer, error) {
	if amount == 0 {
		return BridgeTransfer{}, fmt.Errorf("amount must be positive")
	}
	if _, err := GetBridge(bridgeID); err != nil {
		return BridgeTransfer{}, err
	}
	escrow := ModuleAddress("bridge:" + bridgeID)
	if err := Transfer(ctx, asset, ctx.Caller, escrow, amount); err != nil {
		return BridgeTransfer{}, err
	}
	bt := BridgeTransfer{
		ID:       uuid.New().String(),
		BridgeID: bridgeID,
		From:     ctx.Caller,
		To:       to,
		Asset:    asset,
		Amount:   amount,
		Time:     time.Now().UTC(),
	}
	raw, _ := json.Marshal(bt)
	if err := CurrentStore().Set([]byte("crosschain:transfer:"+bt.ID), raw); err != nil {
		_ = Transfer(ctx, asset, escrow, ctx.Caller, amount)
		return BridgeTransfer{}, err
	}
	Broadcast("bridge:transfer:new", raw)
	return bt, nil
}

// CompleteBridgeTransfer releases locked assets after proof verification.
func CompleteBridgeTransfer(ctx *Context, id string, proof Proof) error {
	bt, err := GetBridgeTransfer(id)
	if err != nil {
		return err
	}
	if bt.Completed {
		return fmt.Errorf("transfer already completed")
	}
	if !verifySPV(proof) {
		return ErrInvalidProof
	}
	escrow := ModuleAddress("bridge:" + bt.BridgeID)
	if err := Transfer(ctx, bt.Asset, escrow, bt.To, bt.Amount); err != nil {
		return err
	}
	bt.Completed = true
	raw, _ := json.Marshal(bt)
	if err := CurrentStore().Set([]byte("crosschain:transfer:"+bt.ID), raw); err != nil {
		return err
	}
	Broadcast("bridge:transfer:complete", raw)
	return nil
}

// GetBridgeTransfer fetches a transfer record by ID.
func GetBridgeTransfer(id string) (BridgeTransfer, error) {
	raw, err := CurrentStore().Get([]byte("crosschain:transfer:" + id))
	if err != nil {
		return BridgeTransfer{}, ErrNotFound
	}
	var bt BridgeTransfer
	if err := json.Unmarshal(raw, &bt); err != nil {
		return BridgeTransfer{}, err
	}
	return bt, nil
}

// ListBridgeTransfers returns all transfer records sorted by creation time.
func ListBridgeTransfers() ([]BridgeTransfer, error) {
	it := CurrentStore().Iterator([]byte("crosschain:transfer:"), nil)
	defer it.Close()
	var out []BridgeTransfer
	for it.Next() {
		var bt BridgeTransfer
		if err := json.Unmarshal(it.Value(), &bt); err != nil {
			return nil, err
		}
		out = append(out, bt)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Time.Before(out[j].Time) })
	return out, it.Error()
}
