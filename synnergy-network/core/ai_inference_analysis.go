package core

// ai_inference_analysis.go - advanced AI inference and transaction analysis
// Provides generic inference execution against registered models and batch
// analysis helpers for transactions. This extends ai.go with additional
// functionality used by consensus and the VM.

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// InferenceResult stores the output of a model inference run along with
// the score returned by the AI service.
type InferenceResult struct {
	Model     [32]byte `json:"model"`
	Output    []byte   `json:"output"`
	Score     float32  `json:"score"`
	Timestamp int64    `json:"time"`
}

// DriftRecord captures a drift event for audit purposes.
type DriftRecord struct {
	Model [32]byte `json:"model"`
	Drift float64  `json:"drift"`
	Time  int64    `json:"time"`
}

// InferModel runs the input through the referenced model and records the result
// on the ledger. The model must have been published previously.
func (ai *AIEngine) InferModel(hash [32]byte, input []byte) (*InferenceResult, error) {
	if ai == nil {
		return nil, errors.New("AI engine not initialised")
	}
	if _, ok := ai.modelMeta(hash); !ok {
		return nil, fmt.Errorf("model %x not found", hash)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := ai.client.Inference(ctx, &TFRequest{Payload: input})
	if err != nil {
		return nil, err
	}
	res := &InferenceResult{Model: hash, Output: resp.Result, Score: resp.Score, Timestamp: time.Now().Unix()}
	key := append([]byte("ai:inference:"), hash[:]...)
	ai.led.SetState(key, mustJSON(res))
	if ai.drift != nil {
		if diff, alert := ai.drift.Record(hash, resp.Score); alert {
			dkey := append([]byte("ai:drift:"), hash[:]...)
			rec := DriftRecord{Model: hash, Drift: diff, Time: time.Now().Unix()}
			ai.led.SetState(dkey, mustJSON(rec))
		}
	}
	return res, nil
}

// AnalyseTransactions runs anomaly detection over a batch of transactions and
// returns a map of tx hash to risk score. It is used by the consensus engine
// and compliance layer to pre-filter blocks.
func (ai *AIEngine) AnalyseTransactions(txs []*Transaction) (map[Hash]float32, error) {
	if ai == nil {
		return nil, errors.New("AI engine not initialised")
	}
	out := make(map[Hash]float32)
	for _, tx := range txs {
		if tx == nil {
			continue
		}
		tx.HashTx()
		score, err := ai.PredictAnomaly(tx)
		if err != nil {
			return out, err
		}
		out[tx.Hash] = score
	}
	return out, nil
}
