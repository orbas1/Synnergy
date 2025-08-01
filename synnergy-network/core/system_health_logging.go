package core

import (
	"encoding/hex"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Metrics captures a snapshot of network and node health statistics.
type Metrics struct {
	Height        uint64 `json:"height"`
	LastHash      string `json:"last_hash"`
	PendingTx     int    `json:"pending_tx"`
	PeerCount     int    `json:"peer_count"`
	TotalSupply   uint64 `json:"total_supply"`
	MemAlloc      uint64 `json:"mem_alloc"`
	NumGoroutines int    `json:"goroutines"`
	Timestamp     int64  `json:"timestamp"`
}

// HealthLogger provides simple system monitoring and structured logging.
type HealthLogger struct {
	ledger  *Ledger
	network *Node
	coin    *Coin
	txpool  *TxPool

	log  *logrus.Logger
	file *os.File
	mu   sync.Mutex
}

// NewHealthLogger configures a HealthLogger writing JSON logs to the given path.
func NewHealthLogger(l *Ledger, n *Node, c *Coin, tp *TxPool, path string) (*HealthLogger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	lg := logrus.New()
	lg.SetFormatter(&logrus.JSONFormatter{})
	lg.SetOutput(f)

	return &HealthLogger{ledger: l, network: n, coin: c, txpool: tp, log: lg, file: f}, nil
}

// Close releases the underlying log file.
func (h *HealthLogger) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.file.Close()
}

// Rotate switches logging to a new file path.
func (h *HealthLogger) Rotate(path string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.file.Close(); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	h.log.SetOutput(f)
	h.file = f
	return nil
}

// LogEvent records an arbitrary message with the specified log level.
func (h *HealthLogger) LogEvent(level logrus.Level, msg string) {
	h.mu.Lock()
	h.log.Log(level, msg)
	h.mu.Unlock()
}

// MetricsSnapshot gathers current metrics from the ledger, network and runtime.
func (h *HealthLogger) MetricsSnapshot() Metrics {
	m := Metrics{Timestamp: time.Now().Unix(), NumGoroutines: runtime.NumGoroutine()}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	m.MemAlloc = mem.Alloc

	if h.ledger != nil {
		m.Height = h.ledger.LastBlockHeight()
		hash := h.ledger.LastBlockHash()
		m.LastHash = hex.EncodeToString(hash[:])
		m.PendingTx = len(h.ledger.TxPool)
	}
	if h.txpool != nil {
		m.PendingTx = len(h.txpool.Snapshot())
	}
	if h.network != nil {
		m.PeerCount = len(h.network.Peers())
	}
	if h.coin != nil {
		m.TotalSupply = h.coin.TotalSupply()
	}
	return m
}
