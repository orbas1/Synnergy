package core

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"math/big"
	"os"
)

// NewLedger initializes a ledger, replaying an existing WAL and optionally loading a genesis block.
func NewLedger(cfg LedgerConfig) (*Ledger, error) {
	// Prepare directories
	// Open or create WAL
	wal, err := os.OpenFile(cfg.WALPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open WAL: %w", err)
	}

	l := &Ledger{
		Blocks:           []*Block{},
		State:            make(map[string][]byte),
		UTXO:             make(map[string]UTXO),
		TxPool:           make(map[string]*Transaction),
		Contracts:        make(map[string]Contract),
		TokenBalances:    make(map[string]uint64),
		walFile:          wal,
		snapshotPath:     cfg.SnapshotPath,
		snapshotInterval: cfg.SnapshotInterval,
	}
	if cfg.GenesisBlock != nil {
		if err := l.applyBlock(cfg.GenesisBlock, false); err != nil {
			return nil, err
		}
		logrus.Infof("Loaded genesis block height %d", cfg.GenesisBlock.Header.Height)
	}
	// Replay WAL
	scanner := bufio.NewScanner(wal)
	for scanner.Scan() {
		var blk Block
		if err := json.Unmarshal(scanner.Bytes(), &blk); err != nil {
			return nil, fmt.Errorf("WAL unmarshal: %w", err)
		}
		if err := l.applyBlock(&blk, false); err != nil {
			return nil, err
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("WAL scan: %w", err)
	}
	return l, nil
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
	fmt.Printf("[EmitApproval] token: %v, owner: %v, spender: %v, amount: %d\n", tokenID, owner, spender, amount)
}

func (l *Ledger) EmitTransfer(tokenID TokenID, from, to Address, amount uint64) {
	fmt.Printf("[EmitTransfer] token: %v, from: %v, to: %v, amount: %d\n", tokenID, from, to, amount)
}

func (l *Ledger) DeductGas(addr Address, amount uint64) {
	fmt.Printf("[DeductGas] from: %v, gas: %d\n", addr, amount)
}

func (l *Ledger) WithinBlock(fn func() error) error {
	// For now, this just wraps the call
	return fn()
}

func (l *Ledger) IsIDTokenHolder(addr Address) bool {
	const IDTokenID TokenID = 1 // Use your actual governance/ID token ID

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
	f.Close()
	// Truncate WAL (start anew)
	l.walFile.Close()
	wal, err := os.Create(l.walFile.Name())
	if err != nil {
		return err
	}
	l.walFile = wal
	logrus.Infof("Snapshot saved to %s; WAL truncated", l.snapshotPath)
	return nil
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
	l.TxPool[fmt.Sprintf("%x", tx.ID)] = tx
	logrus.Infof("Added transaction %x to pool", tx.ID)
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
func (l *Ledger) BalanceOf(address []byte) uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.TokenBalances[fmt.Sprintf("%x", address)]
}

// Snapshot returns JSON state of ledger.
func (l *Ledger) Snapshot() ([]byte, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return json.Marshal(l)
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
	log.Printf("Minted %d of token %s to address %s", amount, tokenID, addr.String())

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
