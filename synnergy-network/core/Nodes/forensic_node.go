package Nodes

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// TransactionLite represents the minimal transaction data required for
// forensic analysis. It mirrors the core.Transaction fields used by the
// node without creating an import cycle.
type TransactionLite interface {
	HashBytes() []byte
}

// LedgerWriter exposes the subset of ledger functionality required by the
// forensic node.
type LedgerWriter interface {
	SetState(key, value []byte)
}

// AIPredictor defines the behaviour required from the AI engine.
type AIPredictor interface {
	PredictAnomaly(tx TransactionLite) (float32, error)
}

// ComplianceEngine defines the minimal compliance checks used by the node.
type ComplianceEngine interface {
	MonitorTransaction(tx TransactionLite, threshold float32) (float32, error)
	StartMonitor(ctx context.Context, txCh <-chan TransactionLite, threshold float32)
}

// ForensicNode provides deep transaction analysis and compliance checks.
type ForensicNode struct {
	node       NodeInterface
	ledger     LedgerWriter
	ai         AIPredictor
	compliance ComplianceEngine
	mu         sync.RWMutex
}

// NewForensicNode wires a forensic node with its ledger and networking facade.
func NewForensicNode(n NodeInterface, led LedgerWriter, ai AIPredictor, comp ComplianceEngine) *ForensicNode {
	return &ForensicNode{node: n, ledger: led, ai: ai, compliance: comp}
}

// AnalyseTransaction runs anomaly detection on the provided transaction
// returning a risk score in the range [0,1]. Results are stored on the ledger
// for audit purposes.
func (f *ForensicNode) AnalyseTransaction(tx TransactionLite) (float32, error) {
	if tx == nil {
		return 0, errors.New("nil transaction")
	}
	if f.ai == nil {
		return 0, errors.New("AI engine not configured")
	}
	score, err := f.ai.PredictAnomaly(tx)
	if err != nil {
		return 0, err
	}
	// store score for forensic audit
	key := append([]byte("forensic:tx:"), tx.HashBytes()...)
	f.mu.Lock()
	if f.ledger != nil {
		f.ledger.SetState(key, []byte(fmt.Sprintf("%f", score)))
	}
	f.mu.Unlock()
	return score, nil
}

// ComplianceCheck validates the transaction against compliance rules and
// returns the anomaly score recorded by the compliance engine.
func (f *ForensicNode) ComplianceCheck(tx TransactionLite, threshold float32) (float32, error) {
	if f.compliance == nil {
		return 0, errors.New("compliance engine not configured")
	}
	score, err := f.compliance.MonitorTransaction(tx, threshold)
	if err != nil {
		return 0, err
	}
	return score, nil
}

// StartMonitoring begins asynchronous monitoring of incoming transactions.
func (f *ForensicNode) StartMonitoring(ctx context.Context, txCh <-chan TransactionLite, threshold float32) {
	if f.compliance == nil {
		return
	}
	f.compliance.StartMonitor(ctx, txCh, threshold)
}
