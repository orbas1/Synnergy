package core

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	Nodes "synnergy-network/core/Nodes"
)

// AIEnhancedConfig aggregates configuration required to start an AI enhanced node.
// It mirrors the BootstrapConfig structure but focuses on ledger and networking.
type AIEnhancedConfig struct {
	Network Config
	Ledger  LedgerConfig
}

// AIEnhancedNode couples a standard network node with a ledger instance and
// exposes AI assisted helpers for predictive analytics. The struct satisfies
// Nodes.AIEnhancedNodeInterface so it can be used without a direct dependency
// on the core package.
type AIEnhancedNode struct {
	*Node
	led    *Ledger
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewAIEnhancedNode creates a node with ledger connectivity. The returned
// instance is ready to be started.
func NewAIEnhancedNode(cfg *AIEnhancedConfig) (*AIEnhancedNode, error) {
	ctx, cancel := context.WithCancel(context.Background())
	n, err := NewNode(cfg.Network)
	if err != nil {
		cancel()
		return nil, err
	}
	led, err := NewLedger(cfg.Ledger)
	if err != nil {
		cancel()
		_ = n.Close()
		return nil, err
	}
	return &AIEnhancedNode{Node: n, led: led, ctx: ctx, cancel: cancel}, nil
}

// Start launches the underlying network services.
func (a *AIEnhancedNode) Start() { go a.ListenAndServe() }

// Stop gracefully shuts down all services.
func (a *AIEnhancedNode) Stop() error {
	a.cancel()
	return a.Close()
}

// Ledger returns the local ledger instance.
func (a *AIEnhancedNode) Ledger() *Ledger { return a.led }

// PredictLoad decodes a JSON encoded slice of TxVolume metrics and delegates
// to the global AI engine for forecasting.
func (a *AIEnhancedNode) PredictLoad(data []byte) (uint64, error) {
	var vol []TxVolume
	if err := json.Unmarshal(data, &vol); err != nil {
		return 0, err
	}
	if AI() == nil {
		return 0, fmt.Errorf("AI engine not initialised")
	}
	return AI().PredictVolume(vol)
}

// AnalyseTx performs anomaly detection over the provided transactions. The
// input is expected to be a JSON array of Transaction objects encoded using the
// core schema. Results are returned as a map of hex encoded tx hashes to risk
// scores.
func (a *AIEnhancedNode) AnalyseTx(data []byte) (map[string]float32, error) {
	var txs []*Transaction
	if err := json.Unmarshal(data, &txs); err != nil {
		return nil, err
	}
	if AI() == nil {
		return nil, fmt.Errorf("AI engine not initialised")
	}
	scores, err := AI().AnalyseTransactions(txs)
	if err != nil {
		return nil, err
	}
	out := make(map[string]float32, len(scores))
	for h, sc := range scores {
		out[hex.EncodeToString(h[:])] = sc
	}
	return out, nil
}

var _ Nodes.AIEnhancedNodeInterface = (*AIEnhancedNode)(nil)
