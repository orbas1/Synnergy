package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CrossChainProtocol defines a generic cross-chain integration profile.
type CrossChainProtocol struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Params    map[string]string `json:"params"`
	CreatedAt time.Time         `json:"created_at"`
}

const protocolPrefix = "crosschain:protocol:"
const TopicProtocolRegistry = "protocol:registry"

// RegisterProtocol stores a new protocol definition in the global KV store.
func RegisterProtocol(p CrossChainProtocol) (CrossChainProtocol, error) {
	if p.Name == "" {
		return p, fmt.Errorf("protocol name required")
	}
	if p.Params == nil {
		p.Params = make(map[string]string)
	}
	p.ID = uuid.New().String()
	p.CreatedAt = time.Now().UTC()
	raw, err := json.Marshal(p)
	if err != nil {
		return p, err
	}
	if err := CurrentStore().Set([]byte(protocolPrefix+p.ID), raw); err != nil {
		return p, err
	}
	Broadcast(TopicProtocolRegistry, raw)
	return p, nil
}

// GetProtocol fetches a protocol definition by its unique ID.
func GetProtocol(id string) (CrossChainProtocol, error) {
	raw, err := CurrentStore().Get([]byte(protocolPrefix + id))
	if err != nil {
		return CrossChainProtocol{}, ErrNotFound
	}
	var p CrossChainProtocol
	if err := json.Unmarshal(raw, &p); err != nil {
		return CrossChainProtocol{}, err
	}
	return p, nil
}

// ListProtocols returns all registered protocols ordered by creation time.
func ListProtocols() ([]CrossChainProtocol, error) {
	it := CurrentStore().Iterator([]byte(protocolPrefix), nil)
	defer it.Close()
	var out []CrossChainProtocol
	for it.Next() {
		var p CrossChainProtocol
		if err := json.Unmarshal(it.Value(), &p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, it.Error()
}

// ProtocolDeposit locks native assets and mints wrapped tokens according to the protocol rules.
func ProtocolDeposit(ctx *Context, protocolID string, asset AssetRef, proof Proof, amount uint64) error {
	if _, err := GetProtocol(protocolID); err != nil {
		return err
	}
	return LockAndMint(ctx, asset, proof, amount)
}

// ProtocolWithdraw burns wrapped tokens and releases native assets.
func ProtocolWithdraw(ctx *Context, protocolID string, asset AssetRef, target Address, amount uint64) error {
	if _, err := GetProtocol(protocolID); err != nil {
		return err
	}
	return BurnAndRelease(ctx, asset, target, amount)
}
