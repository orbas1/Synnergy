package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CrossChainTx records a cross-chain asset movement initiated via LockAndMint
// or BurnAndRelease. Each entry is stored in the shared KV store so external
// services can track bridge usage and audit transfers.
type CrossChainTx struct {
	ID        string    `json:"id"`
	BridgeID  string    `json:"bridge_id"`
	From      Address   `json:"from"`
	To        Address   `json:"to"`
	Asset     AssetRef  `json:"asset"`
	Amount    uint64    `json:"amount"`
	Direction string    `json:"direction"` // "lock_and_mint" or "burn_and_release"
	Proof     Proof     `json:"proof,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// RecordCrossChainTx executes the requested cross-chain operation and persists
// a transaction record in the KV store. It returns the stored entry with a
// populated ID and timestamp.
func RecordCrossChainTx(ctx *Context, tx CrossChainTx) (CrossChainTx, error) {
	logger := zap.L().Sugar()

	if _, err := GetBridge(tx.BridgeID); err != nil {
		logger.Warnf("bridge %s not found: %v", tx.BridgeID, err)
		return CrossChainTx{}, err
	}

	tx.ID = uuid.New().String()
	tx.CreatedAt = time.Now().UTC()
	tx.From = Caller(ctx)

	switch tx.Direction {
	case "lock_and_mint":
		if err := LockAndMint(ctx, tx.Asset, tx.Proof, tx.Amount); err != nil {
			return CrossChainTx{}, err
		}
		tx.To = Caller(ctx)
	case "burn_and_release":
		if err := BurnAndRelease(ctx, tx.Asset, tx.To, tx.Amount); err != nil {
			return CrossChainTx{}, err
		}
	default:
		return CrossChainTx{}, fmt.Errorf("invalid direction")
	}

	raw, err := json.Marshal(tx)
	if err != nil {
		return CrossChainTx{}, err
	}
	key := fmt.Sprintf("crosschain:tx:%s", tx.ID)
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		return CrossChainTx{}, err
	}
	Broadcast("crosschain:tx", raw)
	return tx, nil
}

// GetCrossChainTx fetches a transaction record by ID.
func GetCrossChainTx(id string) (CrossChainTx, error) {
	raw, err := CurrentStore().Get([]byte(fmt.Sprintf("crosschain:tx:%s", id)))
	if err != nil {
		return CrossChainTx{}, ErrNotFound
	}
	var tx CrossChainTx
	if err := json.Unmarshal(raw, &tx); err != nil {
		return CrossChainTx{}, err
	}
	return tx, nil
}

// ListCrossChainTx returns all recorded cross-chain transfers.
func ListCrossChainTx() ([]CrossChainTx, error) {
	it := CurrentStore().Iterator([]byte("crosschain:tx:"), nil)
	defer it.Close()

	var out []CrossChainTx
	for it.Next() {
		var tx CrossChainTx
		if err := json.Unmarshal(it.Value(), &tx); err != nil {
			return nil, err
		}
		out = append(out, tx)
	}
	return out, it.Error()
}
