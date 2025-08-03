package core

// common_structs.go – centralised struct definitions referenced across modules.
// This file **declares only data structures** (no functions) to avoid cyclic
// imports.  Each struct fields reference concrete types from packages imported
// below; to keep this file dependency‑light, we alias external packages with
// minimal scope.
// -----------------------------------------------------------------------------

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
	// Logging & P2P
	"github.com/ethereum/go-ethereum/accounts/abi"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	host "github.com/libp2p/go-libp2p/core/host"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

//---------------------------------------------------------------------
// Coin (minting cap manager)
//---------------------------------------------------------------------

type Coin struct {
	mu          sync.Mutex
	ledger      *Ledger
	totalMinted uint64
}

//---------------------------------------------------------------------
// AI structs (from ai.go)
//---------------------------------------------------------------------

// AIEngine manages AI-related fraud detection and optimization.
type AIEngine struct {
	led    StateRW
	conn   *grpc.ClientConn
	client AIStubClient // manually defined interface
	mu     sync.RWMutex
	models map[[32]byte]ModelMeta
	jobs   map[string]TrainingJob
	encKey []byte        // symmetric key for encrypted storage
	drift  *DriftMonitor // tracks model performance drift
}

type ModelMeta struct {
	CID       string    `json:"cid"`
	Creator   Address   `json:"creator"`
	RoyaltyBp uint16    `json:"royalty_bp"`
	LoadedAt  time.Time `json:"loaded"`
}

//---------------------------------------------------------------------
// AMM edge for path‑finding
//---------------------------------------------------------------------

type edge struct {
	pid    PoolID
	tokenA TokenID
	tokenB TokenID
	price  float64 // mid‑price heuristic
}

//---------------------------------------------------------------------
// Authority subsystem
//---------------------------------------------------------------------

type AuthorityNode struct {
	Addr Address `json:"addr"`
	// Wallet holds the payment address associated with the authority
	// node. It may differ from the node's network address and is used
	// when distributing fees or processing on-chain payments.
	Wallet      Address       `json:"wallet"`
	Role        AuthorityRole `json:"role"`
	Active      bool          `json:"active"`
	PublicVotes uint32        `json:"pv"`
	AuthVotes   uint32        `json:"av"`
	CreatedAt   int64         `json:"since"`
}

type AuthoritySet struct {
	logger  *log.Logger
	led     StateRW
	mu      sync.RWMutex
	members map[Address]struct{}
}

//---------------------------------------------------------------------
// Charity pool
//---------------------------------------------------------------------

type CharityRegistration struct {
	Addr      Address         `json:"addr"`
	Name      string          `json:"name"`
	Category  CharityCategory `json:"cat"`
	Cycle     uint64          `json:"cycle"`
	VoteCount uint32          `json:"votes"`
}

type CharityPool struct {
	mu     sync.Mutex
	logger *log.Logger
	led    StateRW
	vote   Voter

	genesis   time.Time
	lastDaily int64
}

//---------------------------------------------------------------------
// Compliance (KYC) doc
//---------------------------------------------------------------------

type KYCDocument struct {
	Address     Address  `json:"addr"`
	CountryCode string   `json:"cc"`
	IDHash      [32]byte `json:"id_hash"`
	IssuerPK    []byte   `json:"issuer_pk"`
	Signature   []byte   `json:"sig"`
	IssuedAt    int64    `json:"ts"`
}

//---------------------------------------------------------------------
// Consensus core structures (simplified view)
//---------------------------------------------------------------------

type SynnergyConsensus struct {
	logger *log.Logger // or *log.Logger—whichever you use

	ledger *Ledger // ← pointer, not value
	p2p    interface{}
	crypto interface{}
	pool   txPool
	auth   interface{}

	mu            sync.Mutex
	nextSubHeight uint64
	nextBlkHeight uint64
	curDifficulty *big.Int
	blkTimes      []int64

	weights   ConsensusWeights
	weightCfg WeightConfig
}

// ConsensusWeights reflects the active weighting across PoW, PoS and PoH.
// Values are expressed as fractions summing to 1.0.
type ConsensusWeights struct {
	PoW float64
	PoS float64
	PoH float64
}

// WeightConfig groups the coefficients and maximum observed values used when
// computing dynamic consensus weights.
type WeightConfig struct {
	Alpha float64 // network demand weighting
	Beta  float64 // stake concentration weighting
	Gamma float64 // scaling factor per mechanism
	DMax  float64 // maximum observed network demand
	SMax  float64 // maximum observed stake concentration
}

type BlockHeader struct {
	Height    uint64
	Timestamp int64
	PrevHash  []byte
	PoWHash   []byte
	Nonce     uint64
	MinerPk   []byte
}

type SubBlockHeader struct {
	Height    uint64
	Timestamp int64
	Validator []byte
	PoHHash   []byte
	Sig       []byte
}

type SubBlockBody struct{ Transactions [][]byte }

type BlockBody struct{ SubHeaders []SubBlockHeader }

type SubBlock struct {
	Header SubBlockHeader
	Body   SubBlockBody
}

type Block struct {
	Header       BlockHeader    `json:"header"`
	Body         BlockBody      `json:"body"`
	Transactions []*Transaction `json:"txs"` // full ordered list of txs
}

//---------------------------------------------------------------------
// Smart‑contract registry structs
//---------------------------------------------------------------------

type SmartContract struct {
	Address   Address
	Creator   Address
	CodeHash  [32]byte
	Bytecode  []byte
	GasLimit  uint64
	CreatedAt time.Time
}

type RicardianContract struct {
	Address      Address   `json:"address"`
	Version      string    `json:"version"`
	Title        string    `json:"title"`
	Parties      []string  `json:"parties"`
	LegalProse   string    `json:"legal"`
	CodeHash     string    `json:"code_hash"`
	Jurisdiction string    `json:"jurisdiction"`
	Created      time.Time `json:"created"`
}

type ContractRegistry struct {
	*Registry
	ledger *Ledger
	vm     VM
	mu     sync.RWMutex
	byAddr map[Address]*SmartContract
}

//---------------------------------------------------------------------
// Network health checker structs
//---------------------------------------------------------------------

type peerStat struct {
	EWMA       float64
	Misses     int
	LastUpdate time.Time
}

type HealthChecker struct {
	mu        sync.RWMutex
	peers     map[Address]*peerStat
	interval  time.Duration
	alpha     float64
	maxRTT    float64
	maxMisses int
	ping      Pinger
	changer   ViewChanger
	stop      chan struct{}
}

type PeerInfo struct {
	Address Address `json:"address"`
	RTT     float64 `json:"rtt_ms"`
	Misses  int     `json:"misses"`
	Updated int64   `json:"updated_unix"`
}

//---------------------------------------------------------------------
// GreenTech structs (summaries + engine)
//---------------------------------------------------------------------

type UsageRecord struct {
	Validator Address `json:"validator"`
	EnergyKWh float64 `json:"energy_kwh"`
	CarbonKg  float64 `json:"carbon_kg"`
	Timestamp int64   `json:"ts"`
}

type OffsetRecord struct {
	Validator Address `json:"validator"`
	OffsetKg  float64 `json:"offset_kg"`
	Timestamp int64   `json:"ts"`
}

type Certificate string

type nodeSummary struct {
	Energy  float64
	Emitted float64
	Offset  float64
	Score   float64
	Cert    Certificate
}

// CertificateInfo provides a consumable view of a validator's sustainability
// certificate and score. It is used by API callers and CLI tooling to present
// network‑wide environmental metrics.
type CertificateInfo struct {
	Address Address     `json:"address"`
	Score   float64     `json:"score"`
	Cert    Certificate `json:"cert"`
}

type GreenTechEngine struct {
	led StateRW
	mu  sync.RWMutex
}

//---------------------------------------------------------------------
// Ledger core (subset for structs)
//---------------------------------------------------------------------

type LedgerConfig struct {
	GenesisBlock     *Block
	WALPath          string
	SnapshotPath     string
	SnapshotInterval int
	ArchivePath      string // optional gzip file to archive pruned blocks
	PruneInterval    int    // number of recent blocks to retain in memory/WAL
}

// UTXO represents a spendable output identified by (TxID, Index).
type UTXO struct {
	TxID   Hash
	Index  uint32
	Output TxOutput
}

// Contract represents a deployed smart-contract on the Synnergy chain.
type Contract struct {
	//-----------------------------------------------------------------
	// Canonical identifiers
	//-----------------------------------------------------------------
	Address      Address `json:"address"`      // 20-byte runtime address
	DeployTxHash Hash    `json:"deploy_tx"`    // tx that created it
	DeployBlock  uint64  `json:"deploy_block"` // block height at deployment

	//-----------------------------------------------------------------
	// Runtime artefacts
	//-----------------------------------------------------------------
	Bytecode []byte  `json:"bytecode"` // raw EVM (or WASM) code
	ABI      abi.ABI `json:"abi"`      // go-ethereum ABI object

	//-----------------------------------------------------------------
	// Enriched metadata (optional but useful)
	//-----------------------------------------------------------------
	Meta ContractMetadata `json:"meta"`
}

// ContractMetadata stores descriptive & provenance data that
// does *not* affect consensus but is handy for explorers / tooling.
type ContractMetadata struct {
	Name        string    `json:"name"`         // human-readable label
	Version     string    `json:"version"`      // semver or git tag
	Compiler    string    `json:"compiler"`     // e.g. "solc 0.8.23"
	Language    string    `json:"language"`     // "Solidity", "Vyper", …
	SourceHash  Hash      `json:"source_hash"`  // keccak of full source tree
	License     string    `json:"license"`      // SPDX identifier
	Author      string    `json:"author"`       // optional contact
	DocURL      string    `json:"doc_url"`      // off-chain docs / README
	PublishedAt time.Time `json:"published_at"` // UTC timestamp
	Tags        []string  `json:"tags"`         // free-form labels
}

type Ledger struct {
	mu               sync.RWMutex
	Blocks           []*Block
	blockIndex       map[Hash]*Block
	State            map[string][]byte
	UTXO             map[string]UTXO
	TxPool           map[string]*Transaction
	Contracts        map[string]Contract
	TokenBalances    map[string]uint64
	logs             []*Log
	walFile          *os.File
	snapshotPath     string
	snapshotInterval int
	archivePath      string // destination file for archived blocks
	pruneInterval    int    // retain this many recent blocks
	tokens           map[TokenID]Token
	lpBalances       map[Address]map[PoolID]uint64
	nonces           map[Address]uint64
	// NodeLocations stores optional geolocation metadata for peers.
	NodeLocations    map[NodeID]Location
	pendingSubBlocks []SubBlock // <- store sub-blocks here
	holoData         map[Hash][]byte
}

//---------------------------------------------------------------------
// AMM pool & manager
//---------------------------------------------------------------------

type reserve struct {
	token TokenID
	bal   uint64
}

type Pool struct {
	ID      PoolID
	tokenA  TokenID
	tokenB  TokenID
	resA    uint64
	resB    uint64
	totalLP uint64
	feeBps  uint16
	mu      sync.RWMutex
}

type AMM struct {
	logger *log.Logger
	ledger StateRW
	pools  map[PoolID]*Pool
	mu     sync.RWMutex
	nextID PoolID
}

//---------------------------------------------------------------------
// P2P structs
//---------------------------------------------------------------------

type NodeID string

type Peer struct {
	ID      NodeID
	Addr    string
	Latency time.Duration
	Conn    net.Conn
}

type Message struct {
	From  NodeID
	Topic string
	Data  []byte
}

type Config struct {
	ListenAddr     string
	BootstrapPeers []string
	DiscoveryTag   string
}

type Node struct {
	host      host.Host
	pubsub    *pubsub.PubSub
	topics    map[string]*pubsub.Topic
	subs      map[string]*pubsub.Subscription
	topicLock sync.RWMutex
	subLock   sync.RWMutex
	peerLock  sync.RWMutex
	peers     map[NodeID]*Peer
	nat       *NATManager
	ctx       context.Context
	cancel    context.CancelFunc
	cfg       Config
}

//---------------------------------------------------------------------
// Replication
//---------------------------------------------------------------------

// Replicator holds runtime state.
type Replicator struct {
	logger  *log.Logger // ← logrus, not std-lib log
	cfg     *ReplicationConfig
	ledger  BlockReader
	pm      PeerManager
	closing chan struct{}
	wg      sync.WaitGroup
	rangeCh chan []*Block
}

//---------------------------------------------------------------------
// Roll-up structs
//---------------------------------------------------------------------

type BatchHeader struct {
	BatchID   uint64   `json:"id"`
	ParentID  uint64   `json:"parent"`
	TxRoot    [32]byte `json:"tx_root"`
	StateRoot [32]byte `json:"state_root"`
	Submitter Address  `json:"submitter"`
	Timestamp int64    `json:"ts"`
}

type FraudProof struct {
	BatchID   uint64   `json:"id"`
	TxIndex   uint32   `json:"tx_idx"`
	Proof     [][]byte `json:"merkle_proof"`
	Reason    string   `json:"reason"`
	Submitter Address  `json:"submitter"`
}

type Aggregator struct {
	led    StateRW
	mu     sync.Mutex
	nextID uint64
	paused bool
}

//---------------------------------------------------------------------
// Sharding & cross‑shard structs
//---------------------------------------------------------------------

type CrossShardTx struct {
	From      Address `json:"from"`
	To        Address `json:"to"`
	Value     uint64  `json:"value"`
	Payload   []byte  `json:"payload"`
	FromShard ShardID `json:"from_shard"`
	ToShard   ShardID `json:"to_shard"`
	Nonce     uint64  `json:"nonce"`
	Hash      Hash    `json:"hash"`
}

type ShardCoordinator struct {
	led     StateRW
	net     Broadcaster
	mu      sync.RWMutex
	leaders map[ShardID]Address
	metrics map[ShardID]*ShardMetrics
}

//---------------------------------------------------------------------
// Side‑chains bridge structs
//---------------------------------------------------------------------

type Sidechain struct {
	ID         SidechainID `json:"id"`
	Name       string      `json:"name"`
	Threshold  uint8       `json:"threshold"`
	Validators [][]byte    `json:"validators"`
	LastHeight uint64      `json:"last_height"`
	LastRoot   [32]byte    `json:"last_state_root"`
	Paused     bool        `json:"paused"`
	Registered int64       `json:"registered_unix"`
}

type SidechainHeader struct {
	ChainID   SidechainID `json:"chain_id"`
	Height    uint64      `json:"height"`
	Parent    [32]byte    `json:"parent"`
	StateRoot [32]byte    `json:"state_root"`
	TxRoot    [32]byte    `json:"tx_root"`
	SigAgg    []byte      `json:"agg_sig"`
	Timestamp int64       `json:"ts"`
}

//---------------------------------------------------------------------
// Bridge receipts & proofs
//---------------------------------------------------------------------

type DepositReceipt struct {
	Nonce     uint64      `json:"nonce"`
	ChainID   SidechainID `json:"chain"`
	From      Address     `json:"from"`
	To        []byte      `json:"to"`
	Amount    uint64      `json:"amount"`
	Token     TokenID     `json:"token"`
	Timestamp int64       `json:"ts"`
	Hash      [32]byte    `json:"hash"`
}

//---------------------------------------------------------------------
// State‑channel structs
//---------------------------------------------------------------------

type Channel struct {
	ID       ChannelID `json:"id"`
	PartyA   Address   `json:"a"`
	PartyB   Address   `json:"b"`
	ShardA   ShardID   `json:"shard_a"`
	ShardB   ShardID   `json:"shard_b"`
	Token    TokenID   `json:"token"`
	BalanceA uint64    `json:"bal_a"`
	BalanceB uint64    `json:"bal_b"`
	Nonce    uint64    `json:"nonce"`
	Closing  int64     `json:"closing_ts"`
	Paused   bool      `json:"paused"`
}

type SignedState struct {
	Channel Channel `json:"channel"`
	PubKeyA []byte  `json:"pub_key_a"`
	PubKeyB []byte  `json:"pub_key_b"`
	SigA    []byte  `json:"sig_a"`
	SigB    []byte  `json:"sig_b"`
}

type ChannelEngine struct {
	led StateRW
	mu  sync.RWMutex
}

//---------------------------------------------------------------------
// Storage structs
//---------------------------------------------------------------------

type diskEntry struct {
	path string
	size int64
	at   time.Time
}

type diskLRU struct {
	mu    sync.Mutex
	dir   string
	max   int
	index map[string]*diskEntry
	order []*diskEntry
}

type Storage struct {
	logger      *log.Logger
	cfg         *StorageConfig
	client      *http.Client
	cache       *diskLRU
	ledger      MeteredState
	pinEndpoint string
	getEndpoint string
}

//---------------------------------------------------------------------
// TxPool & transaction structs (aggregated from transactions.go)
//---------------------------------------------------------------------

// TxType enumerates high‑level transaction categories.  Its concrete
// definition and associated constants (e.g. TxPayment, TxReversal) reside in
// tx_types.go to keep this file's scope limited to structural declarations.

type Transaction struct {
	// core fields
	Type             TxType            `json:"type"`
	From             Address           `json:"from"`
	To               Address           `json:"to"`
	Value            uint64            `json:"value"`
	GasLimit         uint64            `json:"gas_limit"`
	GasPrice         uint64            `json:"gas_price"`
	Nonce            uint64            `json:"nonce"`
	Timestamp        int64             `json:"timestamp"`
	Payload          []byte            `json:"payload,omitempty"`
	Private          bool              `json:"private,omitempty"`
	EncryptedPayload []byte            `json:"encrypted_payload,omitempty"`
	AuthSigs         [][]byte          `json:"auth_sigs,omitempty"`
	OriginalTx       Hash              `json:"orig,omitempty"`
	Sig              []byte            `json:"sig"`
	Hash             Hash              `json:"hash"`
	Inputs           []TxInput         `json:"inputs,omitempty"`
	Outputs          []TxOutput        `json:"outputs,omitempty"`
	StateChanges     map[string][]byte `json:"state,omitempty"`
	Contract         *Contract         `json:"contract,omitempty"`
	TokenTransfers   []TokenTransfer   `json:"token_transfers,omitempty"`
}

// HashTx returns a simple SHA-256 hash of the transaction contents.
func (tx *Transaction) HashTx() Hash {
	b, _ := json.Marshal(tx)
	return sha256.Sum256(b)
}

// IDHex returns the transaction hash as a hex string. If the hash has not yet
// been computed, it derives it from the transaction contents to ensure a
// stable identifier.
func (tx *Transaction) IDHex() string {
	if tx == nil {
		return ""
	}

	h := tx.Hash
	if h == (Hash{}) {
		h = tx.HashTx()
	}
	return hex.EncodeToString(h[:])
}

type TxInput struct {
	TxID  Hash   // Originating tx hash
	Index uint32 // Output index in that tx
}

type TxOutput struct {
	Address    Address
	Amount     uint64
	PubKeyHash []byte `json:"pk_hash"` // NEW: 20-byte recipient hash
}

type TokenTransfer struct {
	From   Address
	To     Address
	Token  TokenID
	Amount uint64
}

//---------------------------------------------------------------------
// HD Wallet
//---------------------------------------------------------------------

type HDWallet struct {
	seed        []byte
	masterKey   []byte
	masterChain []byte
	logger      *log.Logger
}

// Address represents a 20‑byte account identifier.
type Address [20]byte

// Hash represents a 32‑byte cryptographic hash.
type Hash [32]byte

// -----------------------------------------------------------------------------
// Ledger state interface – minimal read‑write contract
// -----------------------------------------------------------------------------

type StateIterator interface {
	Next() bool
	Key() []byte
	Value() []byte
	Error() error
}

type StateRW interface {
	GetState(key []byte) ([]byte, error)
	SetState(key, value []byte) error
	DeleteState(key []byte) error
	HasState(key []byte) (bool, error)
	PrefixIterator(prefix []byte) StateIterator
	IsIDTokenHolder(addr Address) bool
	Snapshot(func() error) error
	MintLP(to Address, pool PoolID, amt uint64) error
	Transfer(from, to Address, amount uint64) error
	MintToken(to Address, amount uint64) error
	Burn(Address, uint64) error // <- update this line to match implementation
	BalanceOf(addr Address) uint64
	NonceOf(addr Address) uint64
	BurnLP(from Address, pool PoolID, amt uint64) error
	Get(ns, key []byte) ([]byte, error)
	Set(ns, key, val []byte) error
	Mint(addr Address, amount uint64) error
	GetCode(addr Address) []byte
	GetCodeHash(addr Address) Hash
	AddLog(log *Log)
	CreateContract(caller Address, code []byte, value *big.Int, gas uint64) (Address, []byte, bool, error)
	DelegateCall(from Address, to Address, input []byte, value *big.Int, gas uint64) error
	Call(from Address, to Address, input []byte, value *big.Int, gas uint64) ([]byte, error)
	GetContract(addr Address) (*Contract, error)
	GetToken(tokenID TokenID) (Token, error)
	GetTokenBalance(addr Address, tokenID TokenID) (uint64, error)
	SetTokenBalance(addr Address, tokenID TokenID, amount uint64) error
	GetTokenSupply(tokenID TokenID) (uint64, error)
	CallCode(from, to Address, input []byte, value *big.Int, gas uint64) ([]byte, bool, error)
	CallContract(from, to Address, input []byte, value *big.Int, gas uint64) ([]byte, bool, error)
	StaticCall(from, to Address, input []byte, gas uint64) ([]byte, bool, error)
	SelfDestruct(contract Address, beneficiary Address)
}

// -----------------------------------------------------------------------------
// Replication configuration (node‑level YAML section)
// -----------------------------------------------------------------------------

type ReplicationConfig struct {
	MaxConcurrent  int           `yaml:"max_concurrent"`
	ChunksPerSec   int           `yaml:"chunks_per_sec"`
	RetryBackoff   time.Duration `yaml:"retry_backoff"`
	PeerThreshold  int           `yaml:"peer_threshold"`
	Fanout         uint          // √N gossip fan-out
	RequestTimeout time.Duration // per-block fetch timeout
	SyncBatchSize  uint64        // number of blocks per sync request
}

// -----------------------------------------------------------------------------
// Read‑only block chain access for replication / analytics
// -----------------------------------------------------------------------------

type BlockReader interface {
	GetBlock(height uint64) (*Block, error)
	LastHeight() uint64
	HasBlock(hash Hash) bool                    // true if block is in DB
	BlockByHash(hash Hash) (*Block, error)      // fetch full block
	DecodeBlockRLP(data []byte) (*Block, error) // helper for wire payloads
	ImportBlock(b *Block) error                 // add to canonical chain
}

// -----------------------------------------------------------------------------
// Peer management abstraction (used by replication & consensus)
// -----------------------------------------------------------------------------

type PeerManager interface {
	Peers() []PeerInfo
	Connect(addr string) error
	Disconnect(id NodeID) error
	Sample(n int) []string
	SendAsync(peerID, proto string, code byte, payload []byte) error
	Subscribe(proto string) <-chan InboundMsg
	Unsubscribe(proto string)
}

// -----------------------------------------------------------------------------
// Storage subsystem configuration & metered state
// -----------------------------------------------------------------------------

type StorageConfig struct {
	CacheDir         string        `yaml:"cache_dir"`
	MaxCacheBytes    uint64        `yaml:"max_cache_bytes"`
	PinEndpoint      string        `yaml:"pin_endpoint"`
	FetchEndpoint    string        `yaml:"fetch_endpoint"`
	Timeout          time.Duration `yaml:"timeout"`
	CacheSizeEntries int           // max # entries in LRU cache
	IPFSGateway      string        // e.g. https://ipfs.infura.io:5001
	GatewayTimeout   time.Duration // per-request HTTP timeout
}

// MeteredState extends StateRW with gas‑charging (or storage rent) logic.

type MeteredState interface {
	StateRW
	Charge(sender Address, gas uint64) error
	ChargeStorageRent(payer Address, bytes int64) error
}

type SidechainCoordinator struct {
	Ledger StateRW
	Net    Broadcaster
	mu     sync.RWMutex
	Nonce  uint64
}

type WithdrawProof struct {
	Header    SidechainHeader `json:"header"`
	TxData    []byte          `json:"tx_data"`
	Proof     [][]byte        `json:"merkle_proof"`
	TxIndex   uint32          `json:"idx"`
	Recipient Address         `json:"recipient"`
}

// Context holds the transaction-level fields (args, caller, origin, etc.).
type TxContext struct {
	BlockHeight uint64
	TxHash      Hash
	Caller      Address
	Timestamp   int64
	Contract    Address
	Value       *big.Int
	GasPrice    uint64
	Stack       *Stack // EVM evaluation stack
	TxOrigin    Address
	CodeHash    []byte
	GasLimit    uint64
	Method      string
	Args        []byte // transaction input data
	Memory      *Memory
	State       StateRW
}

// Stack is a minimal placeholder for the VM stack structure. It stores
// 256-bit words as *big.Int values in a simple slice-backed stack.
// Using a concrete type avoids interface overhead and ensures type safety
// throughout VM execution.
type Stack struct {
	data []*big.Int
}

// Push adds a *big.Int value onto the stack. A nil value will panic to avoid
// ambiguous entries which could mask programming errors during VM execution.
func (s *Stack) Push(v *big.Int) {
	if v == nil {
		panic("nil value pushed to stack")
	}
	s.data = append(s.data, v)
}

// Pop removes and returns the most recently pushed *big.Int. It panics on an
// empty stack or if the stored value is not a *big.Int, ensuring the VM stack
// remains type-safe.
func (s *Stack) Pop() *big.Int {
	if len(s.data) == 0 {
		panic("stack underflow")
	}
	idx := len(s.data) - 1
	raw := s.data[idx]
	s.data = s.data[:idx]
	val, ok := raw.(*big.Int)
	if !ok {
		panic("stack element is not *big.Int")
	}
	return val
}

// Context is an alias used throughout the codebase for TxContext.
type Context = TxContext

// Call delegates to the underlying state to invoke a contract or high level
// function by name. This is a stub implementation used during early
// development and simply returns an error until the VM wiring is completed.
func (ctx *Context) Call(name string) error {
	return fmt.Errorf("call %s not implemented", name)
}

// Gas deducts the given amount from the remaining gas limit and returns an
// error if insufficient gas is available.
func (ctx *Context) Gas(amount uint64) error {
	if ctx.GasLimit < amount {
		return fmt.Errorf("out of gas")
	}
	ctx.GasLimit -= amount
	return nil
}

type Registry struct {
	mu      sync.RWMutex
	Entries map[string][]byte
	tokens  map[TokenID]Token
}

type TxPool struct {
	mu        sync.RWMutex
	ledger    ReadOnlyState
	gasCalc   GasCalculator
	net       *Broadcaster
	lookup    map[Hash]*Transaction
	queue     []*Transaction
	authority *AuthoritySet
}

type ReadOnlyState interface {
	Get(key string) ([]byte, error)
	BalanceOf(addr Address) uint64
	NonceOf(addr Address) uint64
}

type GasCalculator interface {
	Estimate(payload []byte) (uint64, error)
	Calculate(op string, amount uint64) uint64
}

type InboundMsg struct {
	PeerID  string `json:"peer_id"` // sender’s peer-ID
	Code    byte   `json:"code"`    // protocol-level message code
	Payload []byte `json:"payload"` // opaque payload

	Topic string  `json:"topic,omitempty"` // optional pub-sub topic
	From  Address `json:"from,omitempty"`  // optional address
	Ts    int64   `json:"ts"`              // unix-milliseconds timestamp
}

type NetworkMessage struct {
	Source    Address `json:"source"`
	Target    Address `json:"target"`
	MsgType   string  `json:"type"`
	Content   []byte  `json:"content"`
	Timestamp int64   `json:"timestamp"`
	Topic     string
}
