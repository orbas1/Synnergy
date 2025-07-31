// core/storage.go
package core

// Storage subsystem — chunked IPFS / Arweave gateway wrapper with on-disk LRU
// cache.  Thread-safe and gas-aware.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	logrus "github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// -----------------------------------------------------------------------------
// LRU on-disk cache implementation
// -----------------------------------------------------------------------------

const defaultCacheEntries = 10_000

func newDiskLRU(dir string, maxEntries int) (*diskLRU, error) {
	if maxEntries <= 0 {
		maxEntries = defaultCacheEntries
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &diskLRU{
		dir:   dir,
		max:   maxEntries,
		index: make(map[string]*diskEntry),
	}, nil
}

func (l *diskLRU) put(cid string, data []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if ent, ok := l.index[cid]; ok {
		ent.at = time.Now()
		return nil // already cached
	}

	// Evict if full.
	if len(l.index) >= l.max && len(l.order) > 0 {
		oldest := l.order[0]
		_ = os.Remove(oldest.path)
		delete(l.index, filepath.Base(oldest.path))
		l.order = l.order[1:]
	}

	p := filepath.Join(l.dir, cid)
	if err := os.WriteFile(p, data, 0o644); err != nil {
		return err
	}
	ent := &diskEntry{path: p, size: int64(len(data)), at: time.Now()}
	l.index[cid] = ent
	l.order = append(l.order, ent)
	return nil
}

func (l *diskLRU) get(cid string) ([]byte, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	ent, ok := l.index[cid]
	if !ok {
		return nil, false
	}
	ent.at = time.Now()

	b, err := os.ReadFile(ent.path)
	if err != nil {
		return nil, false
	}
	return b, true
}

// -----------------------------------------------------------------------------
// Storage struct
// -----------------------------------------------------------------------------

// NewStorage wires a Storage instance.
func NewStorage(cfg *StorageConfig, lg *logrus.Logger, led MeteredState) (*Storage, error) {
	if cfg == nil {
		return nil, errors.New("storage config nil")
	}
	cache, err := newDiskLRU(cfg.CacheDir, cfg.CacheSizeEntries)
	if err != nil {
		return nil, fmt.Errorf("cache: %w", err)
	}
	s := &Storage{
		logger: lg,
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.GatewayTimeout},
		cache:  cache,
		ledger: led,

		pinEndpoint: cfg.IPFSGateway + "/api/v0/add?pin=true",
		getEndpoint: cfg.IPFSGateway + "/ipfs/", // append CID
	}
	lg.Infof("storage: gateway %s cache %s", cfg.IPFSGateway, cfg.CacheDir)
	return s, nil
}

// -----------------------------------------------------------------------------
// Public API — Pin & Retrieve
// -----------------------------------------------------------------------------

// Pin uploads data to IPFS gateway, returns CID and byte-length.
func (s *Storage) Pin(ctx context.Context, data []byte, payer Address) (string, int64, error) {
	// Compute deterministic CID locally.
	encodedMH, err := mh.Sum(data, mh.SHA2_256, -1)
	if err != nil {
		return "", 0, err
	}
	c := cid.NewCidV1(cid.Raw, encodedMH)
	cidStr := c.String() // ← String() gives lower-case Base32-CIDv1

	// Already cached?
	if _, ok := s.cache.get(cidStr); ok {
		return cidStr, int64(len(data)), nil
	}

	// ----------------- pin via gateway -----------------
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.pinEndpoint, bytes.NewReader(data))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return "", 0, fmt.Errorf("gateway pin %d: %s", resp.StatusCode, string(b))
	}

	var meta struct {
		Hash string `json:"Hash"`
		Size string `json:"Size"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return "", 0, fmt.Errorf("decode: %w", err)
	}
	if meta.Hash != cidStr {
		return "", 0, errors.New("cid mismatch between local and gateway")
	}

	// Cache locally (best-effort).
	_ = s.cache.put(cidStr, data)

	// Charge gas if ledger provided.
	if s.ledger != nil {
		if err := s.ledger.ChargeStorageRent(payer, int64(len(data))); err != nil {
			s.logger.Printf("storage rent charge failed: %v", err)
		}
	}

	s.logger.Printf("pinned CID %s (%d bytes)", cidStr, len(data))
	return cidStr, int64(len(data)), nil
}

// Retrieve returns data for CID (cache → gateway fallback).
func (s *Storage) Retrieve(ctx context.Context, cidStr string) ([]byte, error) {
	if b, ok := s.cache.get(cidStr); ok {
		return b, nil
	}

	url := s.getEndpoint + cidStr
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 128))
		return nil, fmt.Errorf("gateway fetch %d: %s", resp.StatusCode, string(b))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	_ = s.cache.put(cidStr, data) // best-effort

	s.logger.Printf("retrieved CID %s (%d bytes)", cidStr, len(data))
	return data, nil
}

func (s *Storage) vmPin(data []byte, caller Address) (string, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.GatewayTimeout)
	defer cancel()
	return s.Pin(ctx, data, caller)
}

// StorageListing represents a provider's storage offer
type StorageListing struct {
	ID         string    `json:"id"`
	Provider   Address   `json:"provider"`
	PricePerGB uint64    `json:"price_per_gb"`
	CapacityGB int       `json:"capacity_gb"`
	CreatedAt  time.Time `json:"created_at"`
}

// StorageDeal represents a client's purchase or rental deal
type StorageDeal struct {
	ID        string        `json:"id"`
	ListingID string        `json:"listing_id"`
	Client    Address       `json:"client"`
	Duration  time.Duration `json:"duration"`
	EscrowID  string        `json:"escrow_id"`
	CreatedAt time.Time     `json:"created_at"`
	Closed    bool          `json:"closed"`
	ClosedAt  *time.Time    `json:"closed_at,omitempty"`
}

// CreateListing registers a new storage offer
func CreateListing(l *StorageListing) error {
	logger := zap.L().Sugar()
	// Validate provider identity
	if !Exists(l.Provider) {
		return ErrUnauthorized
	}
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	l.CreatedAt = time.Now().UTC()
	key := fmt.Sprintf("storage:listing:%s", l.ID)

	raw, err := json.Marshal(l)
	if err != nil {
		logger.Errorf("marshal listing failed: %v", err)
		return err
	}
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		logger.Errorf("persist listing failed: %v", err)
		return err
	}
	logger.Infof("Storage listing created: %s", l.ID)
	return nil
}

func Exists(addr Address) bool {
	key := []byte(fmt.Sprintf("identity:provider:%x", addr))
	val, err := CurrentStore().Get(key)
	return err == nil && val != nil
}

// OpenDeal creates an escrow-backed storage deal
func OpenDeal(d *StorageDeal) (*Escrow, error) {
	logger := zap.L().Sugar()
	// Validate client identity
	if !Exists(d.Client) {
		return nil, ErrUnauthorized
	}
	// Fetch listing
	listKey := fmt.Sprintf("storage:listing:%s", d.ListingID)
	rawList, err := CurrentStore().Get([]byte(listKey))
	if err != nil {
		return nil, ErrNotFound
	}
	var listing StorageListing
	if err := json.Unmarshal(rawList, &listing); err != nil {
		return nil, err
	}
	// Compute total price
	price := listing.PricePerGB * uint64(listing.CapacityGB)
	// Create escrow: client pays price to provider
	esc, err := Create(listing.Provider, d.Client, price)
	if err != nil {
		logger.Errorf("escrow create failed: %v", err)
		return nil, err
	}
	d.EscrowID = esc.ID
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	d.CreatedAt = time.Now().UTC()
	dealKey := fmt.Sprintf("storage:deal:%s", d.ID)
	rawDeal, err := json.Marshal(d)
	if err != nil {
		logger.Errorf("marshal deal failed: %v", err)
		return nil, err
	}
	if err := CurrentStore().Set([]byte(dealKey), rawDeal); err != nil {
		logger.Errorf("persist deal failed: %v", err)
		return nil, err
	}
	logger.Infof("Storage deal opened: %s", d.ID)
	return esc, nil
}

func Create(provider, client Address, amount uint64) (*Escrow, error) {
	esc := &Escrow{
		ID:     uuid.New().String(),
		Buyer:  client,
		Seller: provider,
		Amount: amount,
		State:  "funded",
	}

	escrowKey := fmt.Sprintf("escrow:%s", esc.ID)
	data, _ := json.Marshal(esc)

	if err := CurrentStore().Set([]byte(escrowKey), data); err != nil {
		return nil, err
	}

	// Optionally: transfer funds from client to module escrow account
	escrowAccount := ModuleAddress("storage_escrow")
	if err := Transfer(nil, AssetRef{Kind: AssetCoin}, client, escrowAccount, amount); err != nil {
		return nil, err
	}

	return esc, nil
}

func CloseDeal(ctx *Context, dealID string) error {
	logger := zap.L().Sugar()
	dealKey := fmt.Sprintf("storage:deal:%s", dealID)
	raw, err := CurrentStore().Get([]byte(dealKey))
	if err != nil {
		return ErrNotFound
	}

	var d StorageDeal
	if err := json.Unmarshal(raw, &d); err != nil {
		return err
	}

	if d.Closed {
		return ErrInvalidState
	}

	// Release escrow
	if err := Release(ctx, d.EscrowID); err != nil {
		logger.Errorf("escrow release failed: %v", err)
		return err
	}

	d.Closed = true
	now := time.Now().UTC()
	d.ClosedAt = &now

	updated, err := json.Marshal(&d)
	if err != nil {
		return err
	}

	if err := CurrentStore().Set([]byte(dealKey), updated); err != nil {
		logger.Errorf("persist deal update failed: %v", err)
		return err
	}

	logger.Infof("Storage deal closed: %s", dealID)
	return nil
}

var (
	ErrInvalidState = errors.New("invalid deal state")
)

func Release(ctx *Context, escrowID string) error {
	key := fmt.Sprintf("escrow:%s", escrowID)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil || raw == nil {
		return fmt.Errorf("escrow not found")
	}

	var esc Escrow
	if err := json.Unmarshal(raw, &esc); err != nil {
		return err
	}

	if esc.State != "funded" {
		return ErrInvalidState
	}

	escrowAccount := ModuleAddress("storage_escrow")
	if err := Transfer(ctx, AssetRef{Kind: AssetCoin}, escrowAccount, esc.Seller, esc.Amount); err != nil {
		return err
	}

	esc.State = "released"
	updated, _ := json.Marshal(esc)
	return CurrentStore().Set([]byte(key), updated)
}
