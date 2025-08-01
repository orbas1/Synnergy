package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ContractMapping links a local contract address to a remote chain address.
type ContractMapping struct {
	Local       Address   `json:"local"`
	RemoteChain string    `json:"remote_chain"`
	Remote      string    `json:"remote"`
	CreatedAt   time.Time `json:"created_at"`
}

const topicContractRegistry = "xchain:contract"

// RegisterContract stores a mapping from a local contract address to a remote chain.
// RegisterXContract stores a mapping from a local contract address to a remote chain.
func RegisterXContract(local Address, remoteChain, remoteAddr string) error {
	logger := zap.L().Sugar()
	key := fmt.Sprintf("crosschain:contract:%s", hex.EncodeToString(local[:]))
	m := ContractMapping{Local: local, RemoteChain: remoteChain, Remote: remoteAddr, CreatedAt: time.Now().UTC()}
	raw, err := json.Marshal(m)
	if err != nil {
		logger.Errorf("marshal contract mapping: %v", err)
		return err
	}
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		logger.Errorf("store contract mapping: %v", err)
		return err
	}
	Broadcast(topicContractRegistry, raw)
	logger.Infof("registered cross-chain contract %x -> %s:%s", local, remoteChain, remoteAddr)
	return nil
}

// GetXContract retrieves a mapping by local address.
func GetXContract(local Address) (ContractMapping, error) {
	raw, err := CurrentStore().Get([]byte(fmt.Sprintf("crosschain:contract:%s", hex.EncodeToString(local[:]))))
	if err != nil {
		return ContractMapping{}, ErrNotFound
	}
	var m ContractMapping
	if err := json.Unmarshal(raw, &m); err != nil {
		return ContractMapping{}, err
	}
	return m, nil
}

// ListContracts returns all registered contract mappings.
// ListXContracts returns all registered contract mappings.
func ListXContracts() ([]ContractMapping, error) {
	it := CurrentStore().Iterator([]byte("crosschain:contract:"), nil)
	defer it.Close()
	var out []ContractMapping
	for it.Next() {
		var m ContractMapping
		if err := json.Unmarshal(it.Value(), &m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, it.Error()
}

// RemoveContract deletes a mapping for the given local address.
// RemoveXContract deletes a mapping for the given local address.
func RemoveXContract(local Address) error {
	return CurrentStore().Delete([]byte(fmt.Sprintf("crosschain:contract:%s", hex.EncodeToString(local[:]))))
}
