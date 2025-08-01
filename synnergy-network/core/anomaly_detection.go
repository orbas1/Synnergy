package core

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// AnomalyService provides anomaly detection helpers that integrate
// the AI engine with the ledger. Detected transactions are persisted
// so other modules such as consensus or smart contracts can query the
// risk status.
type AnomalyService struct {
	ledger    *Ledger
	threshold float32

	mu      sync.RWMutex
	flagged map[Hash]float32
}

// NewAnomalyService creates a new service instance. The ledger pointer is
// required to store flagged transactions. Threshold defines the minimum score
// above which a transaction is marked as anomalous.
func NewAnomalyService(l *Ledger, threshold float32) *AnomalyService {
	return &AnomalyService{
		ledger:    l,
		threshold: threshold,
		flagged:   make(map[Hash]float32),
	}
}

// Analyze runs the AI model on a transaction and flags it if the score
// exceeds the configured threshold. The score is returned to the caller.
func (a *AnomalyService) Analyze(tx *Transaction) (float32, error) {
	if tx == nil {
		return 0, errors.New("nil tx")
	}
	ai := AI()
	if ai == nil {
		return 0, errors.New("AI engine not initialised")
	}
	score, err := ai.PredictAnomaly(tx)
	if err != nil {
		return 0, err
	}
	if score >= a.threshold {
		if err := a.Flag(tx, score); err != nil {
			logrus.WithError(err).Warn("anomaly: flag tx")
		}
	}
	return score, nil
}

// Flag persists the anomaly score of a transaction in the ledger state and
// caches it in memory. Downstream modules can query the ledger to confirm a
// transaction was flagged.
func (a *AnomalyService) Flag(tx *Transaction, score float32) error {
	if tx == nil {
		return errors.New("nil tx")
	}
	h := tx.HashTx()
	a.mu.Lock()
	a.flagged[h] = score
	a.mu.Unlock()
	a.ledger.SetState(a.key(h), []byte(fmt.Sprintf("%.4f", score)))
	return nil
}

// IsFlagged reports whether a transaction hash has been marked as anomalous.
func (a *AnomalyService) IsFlagged(h Hash) bool {
	a.mu.RLock()
	_, ok := a.flagged[h]
	a.mu.RUnlock()
	return ok
}

// Flagged returns a snapshot of all flagged transactions with their scores.
func (a *AnomalyService) Flagged() map[Hash]float32 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make(map[Hash]float32, len(a.flagged))
	for h, s := range a.flagged {
		out[h] = s
	}
	return out
}

func (a *AnomalyService) key(h Hash) []byte {
	return []byte("anomaly:" + hex.EncodeToString(h[:]))
}

//---------------------------------------------------------------------
// Global helpers used by CLI and VM opcodes
//---------------------------------------------------------------------

var (
	anomalyOnce sync.Once
	anomalySvc  *AnomalyService
)

// InitAnomalyService initialises the global anomaly detector using the current
// ledger instance. It is safe to call multiple times but only the first
// invocation has an effect.
func InitAnomalyService(threshold float32) error {
	l := CurrentLedger()
	if l == nil {
		return errors.New("ledger not initialised")
	}
	anomalyOnce.Do(func() {
		anomalySvc = NewAnomalyService(l, threshold)
	})
	return nil
}

// Anomaly returns the globally configured anomaly service or nil if it has not
// been initialised.
func Anomaly() *AnomalyService { return anomalySvc }

// AnalyzeAnomaly is an exported helper that wraps the global service for VM
// integration. It returns zero if the service is not initialised.
func AnalyzeAnomaly(tx *Transaction, threshold float32) (float32, error) {
	svc := Anomaly()
	if svc == nil {
		if err := InitAnomalyService(threshold); err != nil {
			return 0, err
		}
		svc = Anomaly()
	}
	if svc == nil {
		return 0, errors.New("anomaly service not initialised")
	}
	// update threshold dynamically if caller specifies
	if threshold != svc.threshold {
		svc.threshold = threshold
	}
	return svc.Analyze(tx)
}

// FlagAnomalyTx exposes the flagging helper to the VM. It is a no-op if the
// service is not ready.
func FlagAnomalyTx(tx *Transaction, score float32) error {
	svc := Anomaly()
	if svc == nil {
		return errors.New("anomaly service not initialised")
	}
	return svc.Flag(tx, score)
}
