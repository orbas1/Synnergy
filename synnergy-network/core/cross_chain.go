package core

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"crypto/sha256"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Bridge defines parameters for a cross-chain bridge
type Bridge struct {
	ID          string    `json:"id"`
	SourceChain string    `json:"source_chain"`
	TargetChain string    `json:"target_chain"`
	Relayer     Address   `json:"relayer"`
	CreatedAt   time.Time `json:"created_at"`
}

type Proof struct {
	TxHash      []byte
	MerkleRoot  []byte
	MerkleProof [][]byte
	TxIndex     uint32
}

func verifySPV(proof Proof) bool {
	return VerifyMerkleProof(
		proof.TxHash,      // []byte
		proof.MerkleRoot,  // []byte
		proof.MerkleProof, // [][]byte
		proof.TxIndex,     // uint32
	)
}

// RegisterBridge registers a new cross-chain bridge configuration
func RegisterBridge(b Bridge) error {
	logger := zap.L().Sugar()
	relayerHex := hex.EncodeToString(b.Relayer[:]) // convert [20]byte to string
	logger.Infof("Registering bridge from %s to %s by relayer %s", b.SourceChain, b.TargetChain, relayerHex)

	if err := AssertRelayer(relayerHex); err != nil {
		logger.Warnf("Unauthorized relayer %s: %v", relayerHex, err)
		return ErrUnauthorized
	}

	if !HasActiveConnection(b.SourceChain, b.TargetChain) {
		logger.Warnf("No active connection between %s and %s", b.SourceChain, b.TargetChain)
		return fmt.Errorf("no active connection between %s and %s: %w", b.SourceChain, b.TargetChain, ErrNoActiveConnection)
	}

	// assign ID
	b.ID = uuid.New().String()
	b.CreatedAt = time.Now().UTC()
	key := fmt.Sprintf("crosschain:bridge:%s", b.ID)

	raw, err := json.Marshal(b)
	if err != nil {
		logger.Errorf("Failed to marshal bridge: %v", err)
		return err
	}
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		logger.Errorf("Ledger write failed: %v", err)
		return err
	}

	// broadcast new bridge
	Broadcast(TopicBridgeRegistry, raw)
	logger.Infof("Bridge %s registered successfully", b.ID)
	return nil
}

// AssertRelayer checks if the relayer is authorized to perform cross-chain operations
func AssertRelayer(relayer string) error {
	// Add your logic to verify if the relayer is in your authorized list
	if !AuthorizedRelayers[relayer] {
		return fmt.Errorf("relayer %s not authorized", relayer)
	}
	return nil
}

// AuthorizedRelayers is a map of authorized relayer addresses
var AuthorizedRelayers = map[string]bool{
	"relayer1": true,
	"relayer2": true,
}

var ErrUnauthorized = errors.New("unauthorized relayer")

type KVStore interface {
	Set(key, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
	Iterator(start, end []byte) Iterator
}

type Iterator interface {
	Next() bool    // Moves to the next item
	Key() []byte   // Returns the current key
	Value() []byte // Returns the current value
	Error() error  // Returns any iteration error
	Close() error  // Frees resources
}

var (
	storeMu  sync.RWMutex
	appStore KVStore = NewInMemoryStore()
)

// SetStore swaps the global KV store used by cross‑chain utilities. A nil
// argument falls back to an in-memory store. It is safe for concurrent use.
func SetStore(st KVStore) {
	storeMu.Lock()
	defer storeMu.Unlock()
	if st == nil {
		st = NewInMemoryStore()
	}
	appStore = st
}

// CurrentStore returns the globally configured KV store in a thread-safe way.
func CurrentStore() KVStore {
	storeMu.RLock()
	defer storeMu.RUnlock()
	return appStore
}

type InMemoryIterator struct {
	keys   [][]byte
	values [][]byte
	index  int
}

func (it *InMemoryIterator) Next() bool {
	it.index++
	return it.index < len(it.keys)
}

func (it *InMemoryIterator) Key() []byte {
	return it.keys[it.index]
}

func (it *InMemoryIterator) Value() []byte {
	return it.values[it.index]
}

func (it *InMemoryIterator) Error() error {
	return nil
}

func (it *InMemoryIterator) Close() error {
	return nil
}

func (s *InMemoryStore) Iterator(start, end []byte) Iterator {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var keys [][]byte
	var values [][]byte

	for k, v := range s.data {
		key := []byte(k)

		if bytes.HasPrefix(key, start) {
			if end == nil || bytes.Compare(key, end) < 0 {
				keys = append(keys, key)
				values = append(values, v)
			}
		}
	}

	return &InMemoryIterator{
		keys:   keys,
		values: values,
		index:  -1,
	}
}

type InMemoryStore struct {
	data map[string][]byte
	mu   sync.RWMutex
}

// NewInMemoryStore allocates an empty, concurrency-safe KV store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{data: make(map[string][]byte)}
}

func (s *InMemoryStore) Set(key, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data == nil {
		s.data = make(map[string][]byte)
	}
	// store a copy to avoid callers mutating internal state
	v := append([]byte(nil), value...)
	s.data[string(key)] = v
	return nil
}

func (s *InMemoryStore) Get(key []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.data == nil {
		return nil, fmt.Errorf("key not found")
	}
	val, ok := s.data[string(key)]
	if !ok {
		return nil, fmt.Errorf("key not found")
	}
	v := make([]byte, len(val))
	copy(v, val)
	return v, nil
}

var TopicBridgeRegistry = "bridge:registry"

func (s *InMemoryStore) Delete(key []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data == nil {
		return nil
	}
	delete(s.data, string(key))
	return nil
}

// LockAndMint locks native assets from caller and mints wrapped tokens
func LockAndMint(ctx *Context, wrappedAsset AssetRef, proof Proof, amount uint64) error {
	logger := zap.L().Sugar()
	caller := ctx.Caller

	// verify SPV proof
	if !verifySPV(proof) {
		logger.Warnf("SPV proof verification failed for tx %x", proof.TxHash)
		return ErrInvalidProof
	}

	// lock native coin: transfer from caller to escrow
	escrow := ModuleAddress("crosschain")
	if err := Transfer(ctx, AssetRef{Kind: AssetCoin}, caller, escrow, amount); err != nil {
		logger.Errorf("Lock transfer failed: %v", err)
		return err
	}

	// mint wrapped token equal to amount to caller
	if err := Mint(ctx, wrappedAsset, caller, amount); err != nil {
		logger.Errorf("Mint wrapped token failed: %v", err)
		// rollback lock
		_ = Transfer(ctx, AssetRef{Kind: AssetCoin}, escrow, caller, amount)
		return err
	}

	logger.Infof("Locked %d native and minted wrapped to %x", amount, caller)
	return nil
}

// BurnAndRelease burns wrapped tokens and releases native assets to target address
func BurnAndRelease(ctx *Context, wrappedAsset AssetRef, target Address, amount uint64) error {
	logger := zap.L().Sugar()
	caller := ctx.Caller

	// burn wrapped token from caller
	if err := Burn(ctx, wrappedAsset, caller, amount); err != nil {
		logger.Errorf("Burn wrapped token failed: %v", err)
		return err
	}

	// release native coin: transfer from escrow to target
	escrow := ModuleAddress("crosschain")
	if err := Transfer(ctx, AssetRef{Kind: AssetCoin}, escrow, target, amount); err != nil {
		logger.Errorf("Release transfer failed: %v", err)
		// rollback burn by minting back
		_ = Mint(ctx, wrappedAsset, caller, amount)
		return err
	}

	logger.Infof("Burned %d wrapped and released native to %x", amount, target)
	return nil
}

// ListBridges returns all bridge configurations sorted by creation time.
func ListBridges() ([]Bridge, error) {
	it := CurrentStore().Iterator([]byte("crosschain:bridge:"), nil)
	defer it.Close()

	var bridges []Bridge
	for it.Next() {
		var b Bridge
		if err := json.Unmarshal(it.Value(), &b); err != nil {
			return nil, err
		}
		bridges = append(bridges, b)
	}
	sort.Slice(bridges, func(i, j int) bool {
		return bridges[i].CreatedAt.Before(bridges[j].CreatedAt)
	})
	return bridges, it.Error()
}

// GetBridge retrieves a bridge configuration by ID
func GetBridge(id string) (Bridge, error) {
	raw, err := CurrentStore().Get([]byte(fmt.Sprintf("crosschain:bridge:%s", id)))
	if err != nil {
		return Bridge{}, ErrNotFound
	}
	var b Bridge
	if err := json.Unmarshal(raw, &b); err != nil {
		return Bridge{}, err
	}
	return b, nil
}

func Caller(ctx *Context) Address {
	return ctx.Caller
}

var (
	ErrInvalidProof = errors.New("invalid SPV proof")
	ErrNotFound     = errors.New("resource not found")
)

func ModuleAddress(module string) Address {
	hash := sha256.Sum256([]byte("module:" + module))
	return Address(hash[:20]) // cast to your Address type
}
