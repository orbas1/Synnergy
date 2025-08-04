package core_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	core "synnergy-network/core"
	"testing"
	"time"
)

// Mock implementations for StateRW and AIStubClient

type mockLedgerAI struct {
	transfers []string // record From→To strings
	states    map[string][]byte
}

func (m *mockLedgerAI) Transfer(from, to core.Address, amount uint64) error {
	m.transfers = append(m.transfers, from.String()+"->"+to.String())
	return nil
}

func (m *mockLedgerAI) SetState(key, val []byte) {
	if m.states == nil {
		m.states = make(map[string][]byte)
	}
	m.states[string(key)] = val
}

func (m *mockLedgerAI) GetState(key []byte) ([]byte, error) {
	val, ok := m.states[string(key)]
	if !ok {
		return nil, errors.New("not found")
	}
	return val, nil
}

type mockClient struct {
	anomalyResp *core.TFResponse
	anomalyErr  error
	feeResp     *core.TFResponse
	feeErr      error
	volumeResp  *core.TFResponse
	volumeErr   error
}

func (m *mockClient) Anomaly(ctx context.Context, req *core.TFRequest) (*core.TFResponse, error) {
	return m.anomalyResp, m.anomalyErr
}

func (m *mockClient) FeeOpt(ctx context.Context, req *core.TFRequest) (*core.TFResponse, error) {
	return m.feeResp, m.feeErr
}

func (m *mockClient) Volume(ctx context.Context, req *core.TFRequest) (*core.TFResponse, error) {
	return m.volumeResp, m.volumeErr
}

// TestPredictAnomaly covers normal case, nil engine, and error case
func TestPredictAnomaly(t *testing.T) {
	from := core.Address{0x01}
	creator := core.Address{0xAB}
	modelCID := "model123"
	modelHash := sha256.Sum256([]byte(modelCID))

	led := &mockLedgerAI{}
	client := &mockClient{
		anomalyResp: &core.TFResponse{Score: 0.87, Result: []byte(modelCID)},
	}
	ei := &core.AIEngine{led: led, client: client, models: map[[32]byte]core.ModelMeta{
		modelHash: {CID: modelCID, Creator: creator, RoyaltyBp: 100},
	}}

	tx := &core.Transaction{From: from, GasPrice: 10_000}
	score, err := ei.PredictAnomaly(tx)
	if err != nil || score != 0.87 {
		t.Fatalf("unexpected result: %v, %v", score, err)
	}

	if len(led.transfers) != 1 || led.transfers[0] != from.String()+"->"+creator.String() {
		t.Errorf("expected royalty transfer, got %v", led.transfers)
	}

	// case: engine not initialized
	core.ShutdownAI()
	score, err = core.AI().PredictAnomaly(tx)
	if err == nil {
		t.Fatal("expected error when AI engine is nil")
	}
}

func TestOptimizeFees(t *testing.T) {
	expected := uint64(42)
	b, _ := json.Marshal(expected)

	client := &mockClient{
		feeResp: &core.TFResponse{Result: b},
	}
	ei := &core.AIEngine{client: client}

	stats := []core.BlockStats{{GasUsed: 1000, GasLimit: 2000, Interval: time.Second}}
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
	led := &mockLedgerAI{}
	ai := &core.AIEngine{led: led, models: make(map[[32]byte]core.ModelMeta)}

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

func TestPredictVolume(t *testing.T) {
	expected := uint64(99)
	b, _ := json.Marshal(expected)
	client := &mockClient{volumeResp: &core.TFResponse{Result: b}}
	ai := &AIEngine{client: client}
	vols := []core.TxVolume{{Timestamp: time.Now(), Count: 10}}
	val, err := ai.PredictVolume(vols)
	if err != nil || val != expected {
		t.Fatalf("unexpected PredictVolume result: %v %v", val, err)
	}
	client.volumeErr = errors.New("fail")
	client.volumeResp = nil
	_, err = ai.PredictVolume(vols)
	if err == nil {
		t.Fatal("expected error from PredictVolume")
	}
}
