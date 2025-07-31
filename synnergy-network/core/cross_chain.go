package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"sync"
	"time"
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
	Iterator(start, end []byte) Iterator // ← Add this line
}

type Iterator interface {
	Next() bool    // Moves to the next item
	Key() []byte   // Returns the current key
	Value() []byte // Returns the current value
	Error() error  // Returns any iteration error
	Close() error  // Frees resources
}

func CurrentStore() KVStore {
	return appStore // Or wherever you hold the global store instance
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

var appStore = &InMemoryStore{}

type InMemoryStore struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func (s *InMemoryStore) Set(key, value []byte) error {
	s.data[string(key)] = value
	return nil
}
func (s *InMemoryStore) Get(key []byte) ([]byte, error) {
	val, ok := s.data[string(key)]
	if !ok {
		return nil, fmt.Errorf("key not found")
	}
	return val, nil
}

var TopicBridgeRegistry = "bridge:registry"

func (s *InMemoryStore) Delete(key []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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
