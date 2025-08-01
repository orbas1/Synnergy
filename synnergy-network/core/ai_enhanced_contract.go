package core

import (
	"encoding/json"
	"fmt"
	"time"
)

// AIEnhancedContract ties a smart contract to an AI model hash.
type AIEnhancedContract struct {
	ContractAddr Address   `json:"contract"`
	ModelHash    [32]byte  `json:"model_hash"`
	Creator      Address   `json:"creator"`
	DeployedAt   time.Time `json:"deployed_at"`
}

// DeployAIContract compiles and deploys a smart contract while publishing the
// associated AI model. It returns the deployed contract address and model hash.
func DeployAIContract(code []byte, ric []byte, modelCID string, royalty uint16, creator Address, gas uint64) (*AIEnhancedContract, error) {
	if AI() == nil {
		return nil, fmt.Errorf("AI engine not initialised")
	}
	reg := GetContractRegistry()
	if reg == nil {
		return nil, fmt.Errorf("contract registry not initialised")
	}

	// Publish model first so the hash can be stored alongside contract meta.
	modelHash, err := AI().PublishModel(modelCID, creator, royalty)
	if err != nil {
		return nil, err
	}

	addr := DeriveContractAddress(creator, code)
	if err := reg.Deploy(addr, code, ric, gas); err != nil {
		return nil, err
	}

	meta := &AIEnhancedContract{
		ContractAddr: addr,
		ModelHash:    modelHash,
		Creator:      creator,
		DeployedAt:   time.Now().UTC(),
	}

	if led := CurrentLedger(); led != nil {
		b, _ := json.Marshal(meta)
		_ = led.SetState(aiContractKey(addr), b)
	}

	return meta, nil
}

// InvokeAIContract calls a contract method only if the given transaction passes
// the AI fraud prediction threshold. The transaction is provided for scoring and
// must match the invocation arguments.
func InvokeAIContract(ctx *Context, addr Address, method string, args []byte, tx *Transaction, riskThreshold float32) ([]byte, error) {
	if AI() == nil {
		return nil, fmt.Errorf("AI engine not initialised")
	}
	reg := GetContractRegistry()
	if reg == nil {
		return nil, fmt.Errorf("contract registry not initialised")
	}

	score, err := AI().PredictAnomaly(tx)
	if err != nil {
		return nil, err
	}
	if score > riskThreshold {
		return nil, fmt.Errorf("transaction risk %.2f above threshold %.2f", score, riskThreshold)
	}

	return reg.Invoke(ctx.Caller, addr, method, args, ctx.GasLimit)
}

// UpdateAIModel publishes a new model hash and associates it with the contract.
func UpdateAIModel(addr Address, cid string, royalty uint16, creator Address) ([32]byte, error) {
	if AI() == nil {
		return [32]byte{}, fmt.Errorf("AI engine not initialised")
	}

	h, err := AI().PublishModel(cid, creator, royalty)
	if err != nil {
		return [32]byte{}, err
	}

	if led := CurrentLedger(); led != nil {
		b, _ := json.Marshal(map[string][]byte{"model": h[:]})
		_ = led.SetState(aiContractModelKey(addr), b)
	}

	return h, nil
}

// GetAIModel returns the current model hash associated with a contract.
func GetAIModel(addr Address) ([32]byte, error) {
	if led := CurrentLedger(); led != nil {
		raw, err := led.GetState(aiContractModelKey(addr))
		if err == nil && raw != nil {
			var m map[string][]byte
			if jsonErr := json.Unmarshal(raw, &m); jsonErr == nil {
				var h [32]byte
				if v, ok := m["model"]; ok && len(v) == 32 {
					copy(h[:], v)
					return h, nil
				}
			}
		}
	}
	return [32]byte{}, fmt.Errorf("model not found")
}

func aiContractKey(addr Address) []byte {
	return append([]byte("aicontract:"), addr[:]...)
}

func aiContractModelKey(addr Address) []byte {
	return append([]byte("aicontract:model:"), addr[:]...)
}
