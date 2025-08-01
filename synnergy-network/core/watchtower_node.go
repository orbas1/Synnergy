package core

import (
	"context"
	"sync"

	Nodes "synnergy-network/core/Nodes"

	"github.com/sirupsen/logrus"
)

// WatchtowerNode observes transactions and channel updates enforcing contract rules.
type WatchtowerNode struct {
	net    *Node
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
	wt := &WatchtowerNode{
		net:    n,
		ledger: led,
		logger: logrus.New(),
		alerts: make(chan string, 100),
		ctx:    ctx,
		cancel: cancel,
	}
	return wt, nil
}

// Start begins the watch routines and network listener.
func (w *WatchtowerNode) Start() {
	w.mu.Lock()
	defer w.mu.Unlock()
	go w.net.ListenAndServe()
}

// Stop shuts down the node and ledger.
func (w *WatchtowerNode) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.cancel()
	close(w.alerts)
	if err := w.net.Close(); err != nil {
		return err
	}
	return nil
}

// DialSeed proxies to the underlying node.
func (w *WatchtowerNode) DialSeed(seeds []string) error { return w.net.DialSeed(seeds) }

// Broadcast proxies to the underlying node.
func (w *WatchtowerNode) Broadcast(topic string, data []byte) error {
	return w.net.Broadcast(topic, data)
}

// Subscribe proxies to the underlying node.
func (w *WatchtowerNode) Subscribe(topic string) (<-chan []byte, error) {
	return w.net.Subscribe(topic)
}

// ListenAndServe runs the internal network service.
func (w *WatchtowerNode) ListenAndServe() { w.net.ListenAndServe() }

// Close stops all services.
func (w *WatchtowerNode) Close() error { return w.Stop() }

// Peers returns the peer list.
func (w *WatchtowerNode) Peers() []string {
	peers := w.net.Peers()
	out := make([]string, len(peers))
	for i, p := range peers {
		out[i] = string(p.ID)
	}
	return out
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
