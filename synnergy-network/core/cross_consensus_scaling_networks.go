package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CCSNetwork represents a bridge between two independent consensus systems.
type CCSNetwork struct {
	ID              string    `json:"id"`
	SourceConsensus string    `json:"source_consensus"`
	TargetConsensus string    `json:"target_consensus"`
	CreatedAt       time.Time `json:"created_at"`
}

var TopicCCSRegistry = "ccsn:registry"

// RegisterCCSNetwork stores a new cross-consensus configuration and broadcasts it.
func RegisterCCSNetwork(n CCSNetwork) error {
	logger := zap.L().Sugar()
	n.ID = uuid.New().String()
	n.CreatedAt = time.Now().UTC()
	key := fmt.Sprintf("ccsn:network:%s", n.ID)

	raw, err := json.Marshal(n)
	if err != nil {
		logger.Errorf("marshal ccsn: %v", err)
		return err
	}
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		logger.Errorf("store ccsn: %v", err)
		return err
	}

	Broadcast(TopicCCSRegistry, raw)
	logger.Infof("CCSN %s registered", n.ID)
	return nil
}

// ListCCSNetworks returns all registered cross-consensus networks.
func ListCCSNetworks() ([]CCSNetwork, error) {
	it := CurrentStore().Iterator([]byte("ccsn:network:"), nil)
	defer it.Close()

	var nets []CCSNetwork
	for it.Next() {
		var n CCSNetwork
		if err := json.Unmarshal(it.Value(), &n); err != nil {
			return nil, err
		}
		nets = append(nets, n)
	}
	sort.Slice(nets, func(i, j int) bool {
		return nets[i].CreatedAt.Before(nets[j].CreatedAt)
	})
	return nets, it.Error()
}

// GetCCSNetwork retrieves a configuration by ID.
func GetCCSNetwork(id string) (CCSNetwork, error) {
	raw, err := CurrentStore().Get([]byte(fmt.Sprintf("ccsn:network:%s", id)))
	if err != nil {
		return CCSNetwork{}, ErrNotFound
	}
	var n CCSNetwork
	if err := json.Unmarshal(raw, &n); err != nil {
		return CCSNetwork{}, err
	}
	return n, nil
}

// CCSLockAndTransfer locks coin on the current chain and mints wrapped tokens
// on the target consensus network after proof verification.
func CCSLockAndTransfer(ctx *Context, networkID string, wrappedAsset AssetRef, proof Proof, amount uint64) error {
	logger := zap.L().Sugar()
	if !verifySPV(proof) {
		logger.Warnf("SPV proof failed for CCSN tx %x", proof.TxHash)
		return ErrInvalidProof
	}
	caller := ctx.Caller
	escrow := ModuleAddress("ccsn")
	if err := Transfer(ctx, AssetRef{Kind: AssetCoin}, caller, escrow, amount); err != nil {
		logger.Errorf("lock coin: %v", err)
		return err
	}
	if err := Mint(ctx, wrappedAsset, caller, amount); err != nil {
		logger.Errorf("mint wrapped: %v", err)
		_ = Transfer(ctx, AssetRef{Kind: AssetCoin}, escrow, caller, amount)
		return err
	}
	logger.Infof("Locked %d coin and minted wrapped via CCSN", amount)
	return nil
}

// CCSBurnAndRelease burns wrapped tokens and releases coin on this chain.
func CCSBurnAndRelease(ctx *Context, wrappedAsset AssetRef, target Address, amount uint64) error {
	logger := zap.L().Sugar()
	caller := ctx.Caller
	if err := Burn(ctx, wrappedAsset, caller, amount); err != nil {
		logger.Errorf("burn wrapped: %v", err)
		return err
	}
	escrow := ModuleAddress("ccsn")
	if err := Transfer(ctx, AssetRef{Kind: AssetCoin}, escrow, target, amount); err != nil {
		logger.Errorf("release coin: %v", err)
		_ = Mint(ctx, wrappedAsset, caller, amount)
		return err
	}
	logger.Infof("Burned %d wrapped and released coin to %x", amount, target)
	return nil
}

var ErrCCSNNotFound = errors.New("ccsn not found")
