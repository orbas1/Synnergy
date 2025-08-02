package core

// AI module – on‑chain ML hooks for fraud detection, fee optimisation, and model
// storage / monetisation.

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"sync"
	"time"
)

//---------------------------------------------------------------------
// gRPC proto (compiled separately) – minimal stub interface here.
//---------------------------------------------------------------------

type TFRequest struct{ Payload []byte }
type TFResponse struct {
	Score  float32
	Result []byte
}

// ModelUploadRequest uploads a trained model to the remote AI service.
type ModelUploadRequest struct {
	Model []byte
	CID   string
}

// ModelUploadResponse contains an identifier for the uploaded model.
type ModelUploadResponse struct{ ID string }

// TrainingRequest starts a remote training job.
type TrainingRequest struct {
	DatasetCID string
	ModelCID   string
	Params     map[string]string
}

// TrainingResponse returns the remote job identifier.
type TrainingResponse struct{ JobID string }

// TrainingStatusRequest queries the status of a remote training job.
type TrainingStatusRequest struct{ JobID string }

// TrainingStatusResponse reports the status of a remote training job.
type TrainingStatusResponse struct{ Status string }

type AIStubClient interface {
	Anomaly(ctx context.Context, req *TFRequest) (*TFResponse, error)
	FeeOpt(ctx context.Context, req *TFRequest) (*TFResponse, error)
	Volume(ctx context.Context, req *TFRequest) (*TFResponse, error)
	Inference(ctx context.Context, req *TFRequest) (*TFResponse, error)
	Analyse(ctx context.Context, req *TFRequest) (*TFResponse, error)
	UploadModel(ctx context.Context, req *ModelUploadRequest) (*ModelUploadResponse, error)
	StartTraining(ctx context.Context, req *TrainingRequest) (*TrainingResponse, error)
	TrainingStatus(ctx context.Context, req *TrainingStatusRequest) (*TrainingStatusResponse, error)
}

//---------------------------------------------------------------------
// AIEngine singleton
//---------------------------------------------------------------------

var (
	aiOnce sync.Once
	engine *AIEngine
)

func InitAI(led StateRW, grpcEndpoint string, client AIStubClient) error {
	var err error
	aiOnce.Do(func() {
		key := os.Getenv("AI_STORAGE_KEY")
		if len(key) != 32 {
			err = fmt.Errorf("AI_STORAGE_KEY must be 32 bytes")
			return
		}
		conn, e := grpc.Dial(grpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if e != nil {
			err = e
			return
		}
		engine = &AIEngine{
			led:    led,
			conn:   conn,
			client: client,
			models: make(map[[32]byte]ModelMeta),
			jobs:   make(map[string]TrainingJob),
			encKey: []byte(key),
			drift:  NewDriftMonitor(50),
		}
	})
	return err
}

func AI() *AIEngine { return engine }

//---------------------------------------------------------------------
// PredictAnomaly – returns fraud risk [0,1]; triggers compliance if threshold crossed.
//---------------------------------------------------------------------

func (ai *AIEngine) PredictAnomaly(tx *Transaction) (float32, error) {
	if ai == nil {
		return 0, errors.New("AI engine not initialised")
	}

	b, _ := json.Marshal(tx)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := ai.client.Anomaly(ctx, &TFRequest{Payload: b})
	if err != nil {
		return 0, err
	}

	modelHash := sha256.Sum256(resp.Result)
	if meta, ok := ai.modelMeta(modelHash); ok && meta.RoyaltyBp > 0 {
		fee := tx.GasPrice * uint64(meta.RoyaltyBp) / 10_000
		_ = ai.led.Transfer(tx.From, meta.Creator, fee)
	}
	return resp.Score, nil
}

//---------------------------------------------------------------------
// OptimizeFees – predicts optimal basefee based on recent block stats.
//---------------------------------------------------------------------

type BlockStats struct {
	GasUsed  uint64
	GasLimit uint64
	Interval time.Duration
}

// TxVolume represents historical transaction volume metrics used for
// forecasting future network load.
type TxVolume struct {
	Timestamp time.Time
	Count     uint64
}

func (ai *AIEngine) OptimizeFees(stats []BlockStats) (uint64, error) {
	b, _ := json.Marshal(stats)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resp, err := ai.client.FeeOpt(ctx, &TFRequest{Payload: b})
	if err != nil {
		return 0, err
	}

	var target uint64
	if err := json.Unmarshal(resp.Result, &target); err != nil {
		return 0, err
	}
	return target, nil
}

// PredictVolume forecasts the number of transactions expected in the near
// future based on historical volume metrics provided. The returned count can be
// used by consensus or mempool logic to pre-allocate resources.
func (ai *AIEngine) PredictVolume(vol []TxVolume) (uint64, error) {
	if ai == nil {
		return 0, errors.New("AI engine not initialised")
	}
	b, _ := json.Marshal(vol)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resp, err := ai.client.Volume(ctx, &TFRequest{Payload: b})
	if err != nil {
		return 0, err
	}

	var count uint64
	if err := json.Unmarshal(resp.Result, &count); err != nil {
		return 0, err
	}
	return count, nil
}

//---------------------------------------------------------------------
// Model publication & fetch
//---------------------------------------------------------------------

func (ai *AIEngine) PublishModel(cid string, creator Address, royaltyBp uint16) ([32]byte, error) {
	if royaltyBp > 1000 {
		return [32]byte{}, errors.New("royalty exceeds 10% maximum")
	}
	h := sha256.Sum256([]byte(cid))
	meta := ModelMeta{
		CID:       cid,
		Creator:   creator,
		RoyaltyBp: royaltyBp,
		LoadedAt:  time.Now(),
	}
	ai.mu.Lock()
	ai.models[h] = meta
	ai.mu.Unlock()
	ai.led.SetState(modelKey(h), mustJSON(meta))
	return h, nil
}

func (ai *AIEngine) FetchModel(hash [32]byte) (ModelMeta, error) {
	ai.mu.RLock()
	meta, ok := ai.models[hash]
	ai.mu.RUnlock()
	if !ok {
		return ModelMeta{}, errors.New("model not found")
	}
	return meta, nil
}

func (ai *AIEngine) modelMeta(hash [32]byte) (ModelMeta, bool) {
	ai.mu.RLock()
	m, ok := ai.models[hash]
	ai.mu.RUnlock()
	return m, ok
}

//---------------------------------------------------------------------
// Helpers
//---------------------------------------------------------------------

func modelKey(h [32]byte) []byte {
	return append([]byte("ai:model:"), h[:]...)
}

// ModelListing represents an AI model available for sale or rent.
type ModelListing struct {
	ID     string            `json:"id"`
	Seller Address           `json:"seller"`
	Price  uint64            `json:"price"` // price per sale or per hour rental
	Meta   map[string]string `json:"meta"`
}

// Rental holds rental details for an AI model.
type Rental struct {
	ListingID string    `json:"listing_id"`
	Renter    Address   `json:"renter"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
}

// Escrow holds payment until sale/rental conditions are met.
type Escrow struct {
	ID     string  `json:"id"`
	Buyer  Address `json:"buyer"`
	Seller Address `json:"seller"`
	Amount uint64  `json:"amount"`
	State  string  `json:"state"` // "funded", "released"
}

// resolveEscrow finalizes an escrow by releasing funds to the seller.
func resolveEscrow(ctx *Context, e *Escrow) error {
	logger := zap.L().Sugar()

	if e.State != "funded" {
		return fmt.Errorf("escrow %s not in funded state", e.ID)
	}

	escrowAcc := ModuleAddress("ai_marketplace")

	// Transfer funds from escrow account to seller
	if err := Transfer(ctx, AssetRef{Kind: AssetCoin}, escrowAcc, e.Seller, e.Amount); err != nil {
		logger.Errorw("failed to release escrow funds", "escrow", e.ID, "error", err)
		return err
	}

	e.State = "released"

	key := fmt.Sprintf("ai_marketplace:escrow:%s", e.ID)
	data, _ := json.Marshal(e)

	if err := CurrentStore().Set([]byte(key), data); err != nil {
		logger.Errorw("failed to persist escrow state", "escrow", e.ID, "error", err)
		return err
	}

	logger.Infow("escrow released", "escrow", e.ID)
	return nil
}

// ListModel publishes a new model listing.
func ListModel(m *ModelListing) error {
	logger := zap.L().Sugar()

	// KYC check on seller
	if err := ValidateKYC(m.Seller); err != nil {
		return fmt.Errorf("seller KYC failed: %w", err)
	}

	m.ID = uuid.New().String()
	key := fmt.Sprintf("ai_marketplace:listing:%s", m.ID)

	// Check if listing exists
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil {
		return fmt.Errorf("store read error: %w", err)
	}
	if raw != nil {
		return fmt.Errorf("listing %s already exists", m.ID)
	}

	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal listing: %w", err)
	}

	if err := CurrentStore().Set([]byte(key), data); err != nil {
		return fmt.Errorf("persist listing: %w", err)
	}

	logger.Infow("model listed", "id", m.ID, "seller", m.Seller)
	return nil
}

func ValidateKYC(addr Address) error {
	// placeholder logic
	return nil
}

func BuyModel(ctx *Context, listingID string, buyer Address) (*Escrow, error) {
	logger := zap.L().Sugar()

	// Fetch model listing
	key := fmt.Sprintf("ai_marketplace:listing:%s", listingID)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil || raw == nil {
		return nil, fmt.Errorf("listing not found: %w", err)
	}

	var m ModelListing
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("unmarshal listing: %w", err)
	}

	// KYC check on buyer
	if err := ValidateKYC(buyer); err != nil {
		return nil, fmt.Errorf("buyer KYC failed: %w", err)
	}

	// Create escrow object
	esc := &Escrow{
		ID:     uuid.New().String(),
		Buyer:  buyer,
		Seller: m.Seller,
		Amount: m.Price,
		State:  "funded",
	}

	// Transfer funds to escrow account
	escrowAcc := ModuleAddress("ai_marketplace")
	if err := Transfer(ctx, AssetRef{Kind: AssetCoin}, buyer, escrowAcc, m.Price); err != nil {
		return nil, fmt.Errorf("transfer to escrow: %w", err)
	}

	// Persist escrow
	eskKey := fmt.Sprintf("ai_marketplace:escrow:%s", esc.ID)
	data, _ := json.Marshal(esc)
	if err := CurrentStore().Set([]byte(eskKey), data); err != nil {
		return nil, fmt.Errorf("persist escrow: %w", err)
	}

	logger.Infow("model purchase escrow created", "escrow", esc.ID, "buyer", buyer)
	return esc, nil
}

func RentModel(ctx *Context, listingID string, renter Address, duration time.Duration) (*Escrow, error) {
	logger := zap.L().Sugar()

	key := fmt.Sprintf("ai_marketplace:listing:%s", listingID)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil || raw == nil {
		return nil, fmt.Errorf("listing not found: %w", err)
	}

	var m ModelListing
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("unmarshal listing: %w", err)
	}

	if err := ValidateKYC(renter); err != nil {
		return nil, fmt.Errorf("renter KYC failed: %w", err)
	}

	hours := uint64(duration.Hours())
	amount := m.Price * hours

	esc := &Escrow{
		ID:     uuid.New().String(),
		Buyer:  renter,
		Seller: m.Seller,
		Amount: amount,
		State:  "funded",
	}

	escrowAcc := ModuleAddress("ai_marketplace")
	if err := Transfer(ctx, AssetRef{Kind: AssetCoin}, renter, escrowAcc, amount); err != nil {
		return nil, fmt.Errorf("transfer to escrow: %w", err)
	}

	eskKey := fmt.Sprintf("ai_marketplace:escrow:%s", esc.ID)
	data, _ := json.Marshal(esc)
	if err := CurrentStore().Set([]byte(eskKey), data); err != nil {
		return nil, fmt.Errorf("persist escrow: %w", err)
	}

	logger.Infow("model rental escrow created", "escrow", esc.ID, "renter", renter, "duration", duration)
	return esc, nil
}

func ReleaseEscrow(ctx *Context, escrowID string) error {
	eskKey := fmt.Sprintf("ai_marketplace:escrow:%s", escrowID)
	raw, err := CurrentStore().Get([]byte(eskKey))
	if err != nil || raw == nil {
		return fmt.Errorf("escrow not found: %w", err)
	}

	var esc Escrow
	if err := json.Unmarshal(raw, &esc); err != nil {
		return fmt.Errorf("unmarshal escrow: %w", err)
	}

	return resolveEscrow(ctx, &esc)
}
