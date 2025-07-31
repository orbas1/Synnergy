package core

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

// Mock implementations for StateRW and AIStubClient

type mockLedger struct {
	transfers []string // record From→To strings
	states    map[string][]byte
}

func (m *mockLedger) Transfer(from, to Address, amount uint64) error {
	m.transfers = append(m.transfers, from.String()+"->"+to.String())
	return nil
}

func (m *mockLedger) SetState(key, val []byte) {
	if m.states == nil {
		m.states = make(map[string][]byte)
	}
	m.states[string(key)] = val
}

func (m *mockLedger) GetState(key []byte) ([]byte, error) {
	val, ok := m.states[string(key)]
	if !ok {
		return nil, errors.New("not found")
	}
	return val, nil
}

type mockClient struct {
	anomalyResp *TFResponse
	anomalyErr  error
	feeResp     *TFResponse
	feeErr      error
}

func (m *mockClient) Anomaly(ctx context.Context, req *TFRequest) (*TFResponse, error) {
	return m.anomalyResp, m.anomalyErr
}

func (m *mockClient) FeeOpt(ctx context.Context, req *TFRequest) (*TFResponse, error) {
	return m.feeResp, m.feeErr
}

// TestPredictAnomaly covers normal case, nil engine, and error case
func TestPredictAnomaly(t *testing.T) {
	from := Address{0x01}
	creator := Address{0xAB}
	modelCID := "model123"
	modelHash := sha256.Sum256([]byte(modelCID))
	
	led := &mockLedger{}
	client := &mockClient{
		anomalyResp: &TFResponse{Score: 0.87, Result: []byte(modelCID)},
	}
	ei := &AIEngine{led: led, client: client, models: map[[32]byte]ModelMeta{
		modelHash: {CID: modelCID, Creator: creator, RoyaltyBp: 100},
	}}

	tx := &Transaction{From: from, GasPrice: 10_000}
	score, err := ei.PredictAnomaly(tx)
	if err != nil || score != 0.87 {
		t.Fatalf("unexpected result: %v, %v", score, err)
	}

	if len(led.transfers) != 1 || led.transfers[0] != from.String()+"->"+creator.String() {
		t.Errorf("expected royalty transfer, got %v", led.transfers)
	}

	// case: engine not initialized
	engine = nil
	score, err = AI().PredictAnomaly(tx)
	if err == nil {
		t.Fatal("expected error when AI engine is nil")
	}
}

func TestOptimizeFees(t *testing.T) {
	expected := uint64(42)
	b, _ := json.Marshal(expected)
	
	client := &mockClient{
		feeResp: &TFResponse{Result: b},
	}
	ei := &AIEngine{client: client}
	
	stats := []BlockStats{{GasUsed: 1000, GasLimit: 2000, Interval: time.Second}}
	val, err := ei.OptimizeFees(stats)
	if err != nil || val != expected {
		t.Fatalf("unexpected OptimizeFees result: %v %v", val, err)
	}

	client.feeErr = errors.New("fail")
	_, err = ei.OptimizeFees(stats)
	if err == nil {
		t.Fatal("expected error from OptimizeFees")
	}
}

func TestPublishModel(t *testing.T) {
	cid := "abc"
	creator := Address{0xFE}
	led := &mockLedger{}
	ai := &AIEngine{led: led, models: make(map[[32]byte]ModelMeta)}

	hash, err := ai.PublishModel(cid, creator, 99)
	if err != nil {
		t.Fatalf("unexpected publish error: %v", err)
	}

	stored, ok := ai.modelMeta(hash)
	if !ok || stored.Creator != creator || stored.CID != cid {
		t.Fatalf("unexpected model metadata: %+v", stored)
	}

	if len(led.states) == 0 {
		t.Error("expected ledger SetState to be called")
	}

	// error case
	_, err = ai.PublishModel("bad", creator, 5000)
	if err == nil {
		t.Fatal("expected error on high royalty")
	}
}
