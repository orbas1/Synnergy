package core

import (
	"context"
	"sync"

	Nodes "synnergy-network/core/Nodes"

	"github.com/sirupsen/logrus"
)

// WatchtowerNode observes transactions and channel updates enforcing contract rules.
type WatchtowerNode struct {
	*BaseNode
	ledger *Ledger
	logger *logrus.Logger
	alerts chan string
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

type WatchtowerConfig struct {
	Network Config
	Ledger  LedgerConfig
}

// NewWatchtowerNode creates a watchtower node with its own ledger instance.
func NewWatchtowerNode(cfg *WatchtowerConfig) (*WatchtowerNode, error) {
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
	base := NewBaseNode(&NodeAdapter{n})
	wt := &WatchtowerNode{
		BaseNode: base,
		ledger:   led,
		logger:   logrus.New(),
		alerts:   make(chan string, 100),
		ctx:      ctx,
		cancel:   cancel,
	}
	return wt, nil
}

// Start begins the watch routines and network listener.
func (w *WatchtowerNode) Start() {
	w.mu.Lock()
	defer w.mu.Unlock()
	go w.ListenAndServe()
}

// Stop shuts down the node and ledger.
func (w *WatchtowerNode) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.cancel()
	close(w.alerts)
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

// Alerts returns the alert channel.
func (w *WatchtowerNode) Alerts() <-chan string { return w.alerts }

// MonitorTx performs a lightweight fraud check and logs alerts.
func (w *WatchtowerNode) MonitorTx(tx *Transaction) {
	if tx == nil {
		return
	}
	// Placeholder: in real system, analyse tx and enforce contracts.
	if tx.Value == 0 {
		w.alerts <- "zero value transaction detected"
	}
}

// Ensure interface compliance at compile time.
var _ Nodes.NodeInterface = (*WatchtowerNode)(nil)
