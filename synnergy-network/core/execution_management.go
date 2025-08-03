package core

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// ExecutionManager coordinates transaction execution against the ledger
// using the provided VM. It collects executed transactions into blocks and
// finalises them via the ledger. The design intentionally keeps the manager
// lightweight so it can be embedded in CLI tools or consensus services.
//
// This is a minimal implementation used for development and testing. It
// does not perform block validation beyond delegated VM execution.

type ExecutionManager struct {
	mu     sync.Mutex
	ledger *Ledger
	vm     VM

	header BlockHeader
	txs    []*Transaction
}

// NewExecutionManager wires an execution manager with the given ledger and VM.
// Returns nil if the ledger is not provided.
func NewExecutionManager(ledger *Ledger, vm VM) *ExecutionManager {
	if ledger == nil {
		return nil
	}
	return &ExecutionManager{ledger: ledger, vm: vm}
}

// BeginBlock resets internal state and prepares a new block header.
func (em *ExecutionManager) BeginBlock(height uint64) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.header = BlockHeader{Height: height, Timestamp: time.Now().UnixMilli()}
	em.txs = em.txs[:0]
}

// ExecuteTx runs a transaction through the VM and records it if successful.
func (em *ExecutionManager) ExecuteTx(tx *Transaction) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if em.vm == nil {
		return fmt.Errorf("no VM attached")
	}

	ctx := &VMContext{TxContext: TxContext{
		BlockHeight: em.header.Height,
		TxHash:      tx.Hash,
		Caller:      tx.From,
		Timestamp:   em.header.Timestamp,
		GasPrice:    tx.GasPrice,
		GasLimit:    tx.GasLimit,
		Value:       big.NewInt(int64(tx.Value)),
		State:       em.ledger,
	}}

	if _, err := em.vm.Execute(tx.Payload, ctx); err != nil {
		return err
	}
	em.txs = append(em.txs, tx)
	return nil
}

// FinalizeBlock writes the collected transactions to the ledger and returns the
// resulting block structure.
func (em *ExecutionManager) FinalizeBlock() (*Block, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	block := &Block{
		Header:       em.header,
		Transactions: em.txs,
	}
	if err := em.ledger.AddBlock(block); err != nil {
		return nil, err
	}
	return block, nil
}

// MarshalJSON exposes the current block being built. Useful for debugging.
func (em *ExecutionManager) MarshalJSON() ([]byte, error) {
	em.mu.Lock()
	defer em.mu.Unlock()
	type view struct {
		Header BlockHeader    `json:"header"`
		Txs    []*Transaction `json:"txs"`
	}
	return json.Marshal(view{Header: em.header, Txs: em.txs})
}
