package core

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	registry         *prometheus.Registry
	heightGauge      prometheus.Gauge
	pendingTxGauge   prometheus.Gauge
	peerCountGauge   prometheus.Gauge
	totalSupplyGauge prometheus.Gauge
	memAllocGauge    prometheus.Gauge
	goroutinesGauge  prometheus.Gauge
	errorCounter     prometheus.Counter
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
	reg := prometheus.NewRegistry()

	h := &HealthLogger{ledger: l, network: n, coin: c, txpool: tp, log: lg, file: f, registry: reg}

	h.heightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "synnergy_block_height",
		Help: "Current block height of the node",
	})
	h.pendingTxGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "synnergy_pending_transactions",
		Help: "Number of pending transactions",
	})
	h.peerCountGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "synnergy_peer_count",
		Help: "Number of connected peers",
	})
	h.totalSupplyGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "synnergy_total_supply",
		Help: "Total supply of the native coin",
	})
	h.memAllocGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "synnergy_mem_alloc_bytes",
		Help: "Current memory allocation in bytes",
	})
	h.goroutinesGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "synnergy_goroutines",
		Help: "Number of running goroutines",
	})
	h.errorCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "synnergy_log_errors_total",
		Help: "Total number of error events logged",
	})

	reg.MustRegister(
		h.heightGauge,
		h.pendingTxGauge,
		h.peerCountGauge,
		h.totalSupplyGauge,
		h.memAllocGauge,
		h.goroutinesGauge,
		h.errorCounter,
	)

	return h, nil
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
	if level >= logrus.ErrorLevel {
		h.errorCounter.Inc()
	}
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

// RecordMetrics captures the current snapshot and updates Prometheus gauges.
func (h *HealthLogger) RecordMetrics() {
	m := h.MetricsSnapshot()
	h.heightGauge.Set(float64(m.Height))
	h.pendingTxGauge.Set(float64(m.PendingTx))
	h.peerCountGauge.Set(float64(m.PeerCount))
	h.totalSupplyGauge.Set(float64(m.TotalSupply))
	h.memAllocGauge.Set(float64(m.MemAlloc))
	h.goroutinesGauge.Set(float64(m.NumGoroutines))
	h.LogEvent(logrus.InfoLevel, "metrics recorded")
}

// RunMetricsCollector periodically records metrics until the context is canceled.
func (h *HealthLogger) RunMetricsCollector(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			h.RecordMetrics()
		case <-ctx.Done():
			return
		}
	}
}

// StartMetricsServer exposes a Prometheus metrics endpoint on the given address.
// It returns the underlying http.Server so callers may manage its lifecycle.
func (h *HealthLogger) StartMetricsServer(addr string) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(h.registry, promhttp.HandlerOpts{}))
	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			h.LogEvent(logrus.ErrorLevel, err.Error())
		}
	}()
	return srv, nil
}

// ShutdownMetricsServer gracefully stops the metrics HTTP server.
func (h *HealthLogger) ShutdownMetricsServer(ctx context.Context, srv *http.Server) error {
	return srv.Shutdown(ctx)
}
