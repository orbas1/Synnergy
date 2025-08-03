package core

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/sirupsen/logrus"
	"math/big"
	"os"
	"path/filepath"
	"sort"
)

// NewLedger initializes a ledger, replaying an existing WAL and optionally
// loading a genesis block. The WAL file is closed if an error occurs during
// initialisation.
func NewLedger(cfg LedgerConfig) (l *Ledger, err error) {
	// Prepare directories
	// Open or create WAL
	wal, err := os.OpenFile(cfg.WALPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open WAL: %w", err)
	}
	// Ensure the WAL is closed on failure. On success it remains open and is
	// managed by the returned Ledger instance.
	defer func() {
		if err != nil {
			_ = wal.Close()
		}
	}()

	l = &Ledger{
		Blocks:           []*Block{},
		blockIndex:       make(map[Hash]*Block),
		State:            make(map[string][]byte),
		UTXO:             make(map[string]UTXO),
		TxPool:           make(map[string]*Transaction),
		Contracts:        make(map[string]Contract),
		TokenBalances:    make(map[string]uint64),
		lpBalances:       make(map[Address]map[PoolID]uint64),
		nonces:           make(map[Address]uint64),
		NodeLocations:    make(map[NodeID]Location),
		walFile:          wal,
		snapshotPath:     cfg.SnapshotPath,
		snapshotInterval: cfg.SnapshotInterval,
		archivePath:      cfg.ArchivePath,
		pruneInterval:    cfg.PruneInterval,
	}
	if cfg.GenesisBlock != nil {
		if err = l.applyBlock(cfg.GenesisBlock, false); err != nil {
			return nil, err
		}
		logrus.Infof("Loaded genesis block height %d", cfg.GenesisBlock.Header.Height)
	}
	// Replay WAL
	scanner := bufio.NewScanner(wal)
	for scanner.Scan() {
		var blk Block
		if err = json.Unmarshal(scanner.Bytes(), &blk); err != nil {
			return nil, fmt.Errorf("WAL unmarshal: %w", err)
		}
		if err = l.applyBlock(&blk, false); err != nil {
			return nil, err
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("WAL scan: %w", err)
	}
	// initialise global fee distributor for this ledger
	InitTxDistributor(l)
	return l, nil
}

// OpenLedger loads an existing ledger snapshot and replays its WAL. The path
// parameter is treated as a directory containing `ledger.snap` and `ledger.wal`.
// If no snapshot exists, an empty ledger is created.
func OpenLedger(path string) (*Ledger, error) {
	snap := filepath.Join(path, "ledger.snap")
	wal := filepath.Join(path, "ledger.wal")

	var genesis *Block
	l := &Ledger{}

	if f, err := os.Open(snap); err == nil {
		defer f.Close()
		if err := json.NewDecoder(f).Decode(l); err != nil {
			return nil, fmt.Errorf("decode snapshot: %w", err)
		}
		l.snapshotPath = snap
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("open snapshot: %w", err)
	}

	cfg := LedgerConfig{WALPath: wal, SnapshotPath: snap, GenesisBlock: genesis}
	if l.Blocks != nil {
		// ledger restored from snapshot; reuse existing blocks/state
		cfg.GenesisBlock = nil
	}

	loaded, err := NewLedger(cfg)
	if err != nil {
		return nil, err
	}
	if l.Blocks != nil {
		// copy restored data into loaded ledger
		loaded.Blocks = l.Blocks
		loaded.State = l.State
		loaded.UTXO = l.UTXO
		loaded.TxPool = l.TxPool
		loaded.Contracts = l.Contracts
		loaded.TokenBalances = l.TokenBalances
		loaded.NodeLocations = l.NodeLocations
	}
	InitTxDistributor(loaded)
	return loaded, nil
}

func (l *Ledger) GetPendingSubBlocks() []SubBlock {
	l.mu.RLock()
	defer l.mu.RUnlock()

	blocks := make([]SubBlock, len(l.pendingSubBlocks))
	copy(blocks, l.pendingSubBlocks)
	return blocks
}

func (l *Ledger) LastBlockHash() [32]byte {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.Blocks) == 0 {
		return [32]byte{} // empty hash for genesis
	}

	return l.Blocks[len(l.Blocks)-1].Hash()
}

func (l *Ledger) AppendBlock(blk *Block) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Optional: verify hash matches PoW, validate header/prevHash, etc.
	l.Blocks = append(l.Blocks, blk)
	return nil
}

func (l *Ledger) MintBig(addr []byte, amount *big.Int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	key := string(addr) // ⚠️ make sure this is safe as a map key
	if l.TokenBalances == nil {
		l.TokenBalances = make(map[string]uint64)
	}

	l.TokenBalances[key] += amount.Uint64()
}

func (l *Ledger) EmitApproval(tokenID TokenID, owner, spender Address, amount uint64) {
	logrus.WithFields(logrus.Fields{
		"token":   tokenID,
		"owner":   owner,
		"spender": spender,
		"amount":  amount,
	}).Info("EmitApproval")
}

func (l *Ledger) EmitTransfer(tokenID TokenID, from, to Address, amount uint64) {
	logrus.WithFields(logrus.Fields{
		"token":  tokenID,
		"from":   from,
		"to":     to,
		"amount": amount,
	}).Info("EmitTransfer")
}

func (l *Ledger) DeductGas(addr Address, amount uint64) {
	logrus.WithFields(logrus.Fields{
		"from": addr,
		"gas":  amount,
	}).Info("DeductGas")
}

func (l *Ledger) WithinBlock(fn func() error) error {
	// For now, this just wraps the call
	return fn()
}

const IDTokenID TokenID = 1 // Use your actual governance/ID token ID

func (l *Ledger) IsIDTokenHolder(addr Address) bool {
	bal := l.TokenBalance(IDTokenID, addr)
	return bal > 0
}

func (l *Ledger) TokenBalance(tid TokenID, addr Address) uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if token, ok := l.tokens[tid]; ok {
		return token.BalanceOf(addr)
	}
	return 0
}

// applyBlock appends a block and updates sub-ledgers; if persist is true,
// it writes to the WAL and performs snapshots.
func (l *Ledger) applyBlock(block *Block, persist bool) error {
	// 1. Height check
	expected := uint64(len(l.Blocks))
	if block.Header.Height != expected {
		return fmt.Errorf("invalid block height: expected %d, got %d",
			expected, block.Header.Height)
	}

	// 2. Append to canonical chain
	l.Blocks = append(l.Blocks, block)
	h := block.Hash()
	l.blockIndex[h] = block

	// 3. Process each transaction
	for _, tx := range block.Transactions {
		txIDHex := tx.IDHex() // hex string for map keys / logs

		// ---- UTXO updates ---------------------------------------------------
		for _, in := range tx.Inputs {
			key := fmt.Sprintf("%x:%d", in.TxID, in.Index)
			delete(l.UTXO, key)
		}
		for idx, out := range tx.Outputs {
			key := fmt.Sprintf("%x:%d", tx.ID(), idx)
			l.UTXO[key] = UTXO{
				TxID:   tx.ID(),
				Index:  uint32(idx),
				Output: out,
			}
		}

		// ---- State storage updates -----------------------------------------
		for k, v := range tx.StateChanges {
			l.State[k] = v
		}

		// ---- Remove from mem-pool ------------------------------------------
		delete(l.TxPool, txIDHex)

		// ---- Contract deployment -------------------------------------------
		if tx.Contract != nil {
			addrHex := fmt.Sprintf("%x", tx.Contract.Address)
			l.Contracts[addrHex] = *tx.Contract
		}

		// ---- Token transfers -----------------------------------------------
		for _, tr := range tx.TokenTransfers {
			fromHex := fmt.Sprintf("%x", tr.From)
			toHex := fmt.Sprintf("%x", tr.To)
			l.TokenBalances[fromHex] -= tr.Amount
			l.TokenBalances[toHex] += tr.Amount
		}

		// ---- Fee distribution ----------------------------------------
		fee := tx.GasLimit * tx.GasPrice
		dist := CurrentTxDistributor()
		if dist != nil && fee > 0 {
			if err := dist.DistributeFees(tx.From, block.Header.MinerPk, fee); err != nil {
				logrus.Warnf("fee distribution: %v", err)
			}
		}
	}

	// 4. Persistence & snapshots ---------------------------------------------
	if persist {
		data, err := json.Marshal(block)
		if err != nil {
			return fmt.Errorf("marshal block: %w", err)
		}
		if _, err := l.walFile.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("write WAL: %w", err)
		}
		_ = l.walFile.Sync()

		if l.snapshotInterval > 0 && len(l.Blocks)%l.snapshotInterval == 0 {
			if err := l.snapshot(); err != nil {
				logrus.Errorf("snapshot error: %v", err)
			}
		}
		if err := l.prune(); err != nil {
			logrus.Errorf("prune error: %v", err)
		}
	}

	logrus.Infof("Block %d applied; total blocks %d", block.Header.Height, len(l.Blocks))
	return nil
}

// AddBlock is the external entrypoint to append a block.
func (l *Ledger) AddBlock(block *Block) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.applyBlock(block, true)
}

// RebuildChain resets the ledger and replays the supplied blocks as the new
// canonical chain. WAL data is rewritten to reflect the new history. This is
// used during fork recovery to switch to a longer branch.
func (l *Ledger) RebuildChain(blocks []*Block) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Reset core structures
	l.Blocks = make([]*Block, 0, len(blocks))
	l.blockIndex = make(map[Hash]*Block)
	l.State = make(map[string][]byte)
	l.UTXO = make(map[string]UTXO)
	l.TxPool = make(map[string]*Transaction)
	l.Contracts = make(map[string]Contract)
	l.TokenBalances = make(map[string]uint64)
	l.logs = nil
	l.lpBalances = make(map[Address]map[PoolID]uint64)
	l.nonces = make(map[Address]uint64)
	l.NodeLocations = make(map[NodeID]Location)
	l.pendingSubBlocks = nil
	l.holoData = make(map[Hash][]byte)
	l.tokens = make(map[TokenID]Token)

	for i, blk := range blocks {
		if err := l.applyBlock(blk, false); err != nil {
			return fmt.Errorf("reapply block %d: %w", i, err)
		}
	}

	// Rewrite WAL to match new canonical chain
	if l.walFile != nil {
		if err := l.walFile.Truncate(0); err != nil {
			return err
		}
		if _, err := l.walFile.Seek(0, 0); err != nil {
			return err
		}
		enc := json.NewEncoder(l.walFile)
		for _, blk := range l.Blocks {
			if err := enc.Encode(blk); err != nil {
				return err
			}
		}
		_ = l.walFile.Sync()
	}

	InitTxDistributor(l)
	return nil
}

// snapshot writes full ledger state to JSON and truncates WAL.
func (l *Ledger) snapshot() error {
	// Write snapshot
	f, err := os.Create(l.snapshotPath)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(l); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	// Truncate WAL (start anew)
	if err := l.walFile.Close(); err != nil {
		return err
	}
	wal, err := os.Create(l.walFile.Name())
	if err != nil {
		return err
	}
	l.walFile = wal
	logrus.Infof("Snapshot saved to %s; WAL truncated", l.snapshotPath)
	return nil
}

// prune archives old blocks and rewrites WAL to keep the ledger size bounded.
func (l *Ledger) prune() error {
	if l.pruneInterval <= 0 || len(l.Blocks) <= l.pruneInterval {
		return nil
	}

	toArchive := len(l.Blocks) - l.pruneInterval
	if l.archivePath != "" {
		f, err := os.OpenFile(l.archivePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return err
		}
		gz := gzip.NewWriter(f)
		for i := 0; i < toArchive; i++ {
			data, err := json.Marshal(l.Blocks[i])
			if err != nil {
				gz.Close()
				f.Close()
				return err
			}
			if _, err := gz.Write(data); err != nil {
				gz.Close()
				f.Close()
				return err
			}
			if _, err := gz.Write([]byte("\n")); err != nil {
				gz.Close()
				f.Close()
				return err
			}
			delete(l.blockIndex, l.Blocks[i].Hash())
		}
		if err := gz.Close(); err != nil {
			f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}

	l.Blocks = l.Blocks[toArchive:]
	return l.rewriteWAL()
}

// rewriteWAL persists current blocks into WAL from scratch.
func (l *Ledger) rewriteWAL() error {
	if err := l.walFile.Close(); err != nil {
		return err
	}
	wal, err := os.Create(l.walFile.Name())
	if err != nil {
		return err
	}
	l.walFile = wal
	for _, blk := range l.Blocks {
		data, err := json.Marshal(blk)
		if err != nil {
			return err
		}
		if _, err := l.walFile.Write(append(data, '\n')); err != nil {
			return err
		}
	}
	if err := l.walFile.Sync(); err != nil {
		return err
	}
	return nil
}

// StateRoot computes a deterministic hash of the ledger's State map.
func (l *Ledger) StateRoot() Hash {
	l.mu.RLock()
	defer l.mu.RUnlock()

	keys := make([]string, 0, len(l.State))
	for k := range l.State {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha256.New()
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write(l.State[k])
	}
	var out Hash
	copy(out[:], h.Sum(nil))
	return out
}

// GetBlock returns block by height.
func (l *Ledger) GetBlock(height uint64) (*Block, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if height >= uint64(len(l.Blocks)) {
		return nil, fmt.Errorf("block %d not found", height)
	}
	return l.Blocks[height], nil
}

// GetUTXO returns UTXOs for an address.
func (l *Ledger) GetUTXO(address []byte) []UTXO {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var res []UTXO
	addrHex := fmt.Sprintf("%x", address)
	for _, utxo := range l.UTXO {
		if fmt.Sprintf("%x", utxo.Output.PubKeyHash) == addrHex {
			res = append(res, utxo)
		}
	}
	return res
}

// AddToPool adds a transaction to the pool.
func (l *Ledger) AddToPool(tx *Transaction) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.TxPool[fmt.Sprintf("%x", tx.ID())] = tx
	logrus.Infof("Added transaction %x to pool", tx.ID())
}

// ListPool lists pending transactions.
func (l *Ledger) ListPool(limit int) []*Transaction {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var list []*Transaction
	count := 0
	for _, tx := range l.TxPool {
		list = append(list, tx)
		count++
		if limit > 0 && count >= limit {
			break
		}
	}
	return list
}

func (tx *Transaction) ID() Hash {
	return tx.Hash
}

// GetContract returns a deployed contract.
func (l *Ledger) GetContract(address []byte) (*Contract, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	c, ok := l.Contracts[fmt.Sprintf("%x", address)]
	if !ok {
		return nil, fmt.Errorf("contract %x not found", address)
	}
	return &c, nil
}

// BalanceOf returns token balance.
func (l *Ledger) BalanceOf(address Address) uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.TokenBalances[address.String()+":"+Code]
}

// Snapshot returns JSON state of ledger.
func (l *Ledger) Snapshot() ([]byte, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return json.Marshal(l)
}

// HasBlock returns true if the ledger contains a block with the given hash.
func (l *Ledger) HasBlock(h Hash) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, ok := l.blockIndex[h]
	return ok
}

// BlockByHash fetches a block by its hash.
func (l *Ledger) BlockByHash(h Hash) (*Block, error) {
	l.mu.RLock()
	blk, ok := l.blockIndex[h]
	l.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("block %s not found", h.Hex())
	}
	return blk, nil
}

// ImportBlock appends a block to the chain and persists it.
func (l *Ledger) ImportBlock(b *Block) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.applyBlock(b, true)
}

// DecodeBlockRLP decodes an RLP encoded block.
func (l *Ledger) DecodeBlockRLP(data []byte) (*Block, error) {
	var blk Block
	if err := rlp.DecodeBytes(data, &blk); err != nil {
		return nil, err
	}
	return &blk, nil
}

// LastHeight returns the height of the latest block.
func (l *Ledger) LastHeight() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if len(l.Blocks) == 0 {
		return 0
	}
	return l.Blocks[len(l.Blocks)-1].Header.Height
}

// MintToken adds the specified amount to a wallet's balance for a given token.
func (l *Ledger) MintToken(addr Address, tokenID string, amount uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if amount == 0 {
		return fmt.Errorf("mint amount must be positive")
	}

	// Initialize balance map if missing
	if l.TokenBalances == nil {
		l.TokenBalances = make(map[string]uint64)
	}

	key := fmt.Sprintf("%s:%s", addr.String(), tokenID)
	l.TokenBalances[key] += amount

	// Log the minting event (optional if you use structured logging)
	logrus.Infof("Minted %d of token %s to address %s", amount, tokenID, addr.String())

	return nil
}

func (l *Ledger) LastSubBlockHeight() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if len(l.Blocks) == 0 {
		return 0
	}
	return l.Blocks[len(l.Blocks)-1].Header.Height
}

func (l *Ledger) LastBlockHeight() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return uint64(len(l.Blocks) - 1)
}

func (l *Ledger) RecordPoSVote(headerHash []byte, sig []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(headerHash) == 0 || len(sig) == 0 {
		return fmt.Errorf("ledger: empty PoS vote")
	}

	voteKey := fmt.Sprintf("vote:%x", sha256.Sum256(headerHash))
	l.State[voteKey] = sig

	return nil
}

// AppendSubBlock appends a sub-block to the current block-in-progress or ledger.
func (l *Ledger) AppendSubBlock(sb *SubBlock) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Validate sub-block height continuity
	expected := uint64(len(l.Blocks[len(l.Blocks)-1].Body.SubHeaders))
	if sb.Header.Height != expected {
		return fmt.Errorf("ledger: expected sub-block height %d, got %d", expected, sb.Header.Height)
	}

	// Append sub-block header
	l.Blocks[len(l.Blocks)-1].Body.SubHeaders = append(
		l.Blocks[len(l.Blocks)-1].Body.SubHeaders, sb.Header,
	)

	// Optionally append transactions to the pending tx pool or log them
	for _, tx := range sb.Body.Transactions {
		txHash := sha256.Sum256(tx)
		l.TxPool[hex.EncodeToString(txHash[:])] = &Transaction{
			Payload: tx,
			Hash:    txHash,
		}
	}

	return nil
}

func (l *Ledger) Transfer(from, to Address, amount uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.TokenBalances[from.String()] < amount {
		return fmt.Errorf("insufficient balance")
	}

	l.TokenBalances[from.String()] -= amount
	l.TokenBalances[to.String()] += amount
	return nil
}

func (l *Ledger) Mint(to Address, amount uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.TokenBalances[to.String()] += amount
	return nil
}

func (l *Ledger) Burn(from Address, amount uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.TokenBalances[from.String()] < amount {
		return fmt.Errorf("insufficient balance to burn")
	}

	l.TokenBalances[from.String()] -= amount
	return nil
}

// -----------------------------------------------------------------------------
// Additional StateRW helpers
// -----------------------------------------------------------------------------

func (l *Ledger) GetState(key []byte) ([]byte, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	val, ok := l.State[string(key)]
	if !ok {
		return nil, fmt.Errorf("state key not found")
	}
	cpy := make([]byte, len(val))
	copy(cpy, val)
	return cpy, nil
}

func (l *Ledger) SetState(key, value []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	cpy := make([]byte, len(value))
	copy(cpy, value)
	l.State[string(key)] = cpy
	return nil
}

func (l *Ledger) DeleteState(key []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.State, string(key))
	return nil
}

func (l *Ledger) HasState(key []byte) (bool, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, ok := l.State[string(key)]
	return ok, nil
}

type memIter struct {
	keys   [][]byte
	values [][]byte
	idx    int
	err    error
}

func (it *memIter) Next() bool { it.idx++; return it.idx < len(it.keys) }
func (it *memIter) Key() []byte {
	if it.idx < len(it.keys) {
		return it.keys[it.idx]
	}
	return nil
}
func (it *memIter) Value() []byte {
	if it.idx < len(it.values) {
		return it.values[it.idx]
	}
	return nil
}

func (it *memIter) Error() error { return it.err }

func (l *Ledger) PrefixIterator(prefix []byte) StateIterator {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var k [][]byte
	var v [][]byte
	for key, val := range l.State {
		if bytes.HasPrefix([]byte(key), prefix) {
			k = append(k, []byte(key))
			v = append(v, val)
		}
	}
	return &memIter{keys: k, values: v, idx: -1}
}

func (l *Ledger) MintLP(addr Address, pool PoolID, amt uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lpBalances == nil {
		l.lpBalances = make(map[Address]map[PoolID]uint64)
	}
	if l.lpBalances[addr] == nil {
		l.lpBalances[addr] = make(map[PoolID]uint64)
	}
	l.lpBalances[addr][pool] += amt
	return nil
}

func (l *Ledger) BurnLP(addr Address, pool PoolID, amt uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lpBalances == nil || l.lpBalances[addr] == nil {
		return fmt.Errorf("no LP balance")
	}
	if l.lpBalances[addr][pool] < amt {
		return fmt.Errorf("insufficient LP balance")
	}
	l.lpBalances[addr][pool] -= amt
	return nil
}

func (l *Ledger) NonceOf(addr Address) uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.nonces[addr]
}

// AddLog appends an execution log entry to the ledger. The log slice is lazily
// initialised on first use to avoid nil checks across the codebase.
func (l *Ledger) AddLog(log *Log) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.logs == nil {
		l.logs = make([]*Log, 0, 16)
	}
	l.logs = append(l.logs, log)
}

// Call executes a contract located at `to` using the current ledger state as the
// execution context. The call runs inside a transient in-memory state to ensure
// that any side effects are discarded, mirroring the behaviour of an Ethereum
// `eth_call`. It returns the raw bytes produced by the contract or an error if
// execution fails.
func (l *Ledger) Call(from, to Address, input []byte, value *big.Int, gas uint64) ([]byte, error) {
	if l == nil {
		return nil, fmt.Errorf("ledger is nil")
	}
	if value == nil {
		value = big.NewInt(0)
	}

	l.mu.RLock()
	c, ok := l.Contracts[to.String()]
	if !ok {
		l.mu.RUnlock()
		return nil, fmt.Errorf("contract not found at %s", to.String())
	}

	// Make a defensive copy of the contract bytecode so the in-memory call
	// cannot accidentally mutate the original slice stored on the ledger.
	code := append([]byte(nil), c.Bytecode...)

	// Clone the ledger's key/value state to avoid mutating the live ledger.
	stateCopy := make(map[string][]byte, len(l.State))
	for k, v := range l.State {
		stateCopy[k] = append([]byte(nil), v...)
	}

	// Snapshot nonce values so contracts querying account nonces observe a
	// consistent view.
	nonceCopy := make(map[Address]uint64, len(l.nonces))
	for k, v := range l.nonces {
		nonceCopy[k] = v
	}

	// Copy token metadata to satisfy calls that inspect token properties.
	tokenCopy := make(map[TokenID]Token, len(l.tokens))
	for k, v := range l.tokens {
		tokenCopy[k] = v
	}
	l.mu.RUnlock()

	ms := &memState{
		data:       stateCopy,
		balances:   make(map[Address]uint64),
		lpBalances: make(map[Address]map[PoolID]uint64),
		contracts:  map[Address][]byte{to: code},
		tokens:     tokenCopy,
		codeHashes: make(map[Address]Hash),
		nonces:     nonceCopy,
	}

	return ms.Call(from, to, input, value, gas)
}

func (l *Ledger) ChargeStorageRent(addr Address, bytes int64) error {
	if bytes <= 0 {
		return nil
	}
	cost := uint64(bytes)
	zero := AddressZero
	return l.Transfer(addr, zero, cost)
}

// SetNodeLocation stores geolocation information for a node.
func (l *Ledger) SetNodeLocation(id NodeID, loc Location) {
	l.mu.Lock()
	if l.NodeLocations == nil {
		l.NodeLocations = make(map[NodeID]Location)
	}
	l.NodeLocations[id] = loc
	l.mu.Unlock()
}

// GetNodeLocation returns the location for a node if known.
func (l *Ledger) GetNodeLocation(id NodeID) (Location, bool) {
	l.mu.RLock()
	loc, ok := l.NodeLocations[id]
	l.mu.RUnlock()
	return loc, ok
}

// AllNodeLocations returns a copy of the node location table.
func (l *Ledger) AllNodeLocations() map[NodeID]Location {
	l.mu.RLock()
	out := make(map[NodeID]Location, len(l.NodeLocations))
	for id, loc := range l.NodeLocations {
		out[id] = loc
	}
	l.mu.RUnlock()
	return out
}

// Close releases any underlying resources such as the WAL file.
func (l *Ledger) Close() error {
	if l == nil || l.walFile == nil {
		return nil
	}
	return l.walFile.Close()
}
