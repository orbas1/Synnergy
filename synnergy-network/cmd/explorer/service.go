package main

import (
	"encoding/hex"
	"fmt"
	"strings"

	core "synnergy-network/core"
)

// LedgerService wraps common ledger queries used by the Explorer.
type LedgerService struct {
	ledger *core.Ledger
}

func NewLedgerService() (*LedgerService, error) {
	led := core.CurrentLedger()
	if led == nil {
		return nil, fmt.Errorf("ledger not initialised")
	}
	return &LedgerService{ledger: led}, nil
}

// LatestBlocks returns summaries for the most recent blocks.
func (s *LedgerService) LatestBlocks(count int) []map[string]interface{} {
	blocks := s.ledger.Blocks
	if count > len(blocks) {
		count = len(blocks)
	}
	start := len(blocks) - count
	if start < 0 {
		start = 0
	}
	out := make([]map[string]interface{}, 0, count)
	for i := start; i < len(blocks); i++ {
		blk := blocks[i]
		out = append(out, map[string]interface{}{
			"height": blk.Header.Height,
			"hash":   blk.Hash().Hex(),
			"txs":    len(blk.Transactions),
		})
	}
	return out
}

// BlockByHeight returns a block at given height.
func (s *LedgerService) BlockByHeight(h uint64) (*core.Block, error) {
	return s.ledger.GetBlock(h)
}

// TxByID searches for a transaction by hex encoded ID.
func (s *LedgerService) TxByID(hexID string) (*core.Transaction, error) {
	id, err := hex.DecodeString(hexID)
	if err != nil {
		return nil, fmt.Errorf("bad tx id")
	}
	for _, blk := range s.ledger.Blocks {
		for i := range blk.Transactions {
			tx := blk.Transactions[i]
			h := tx.ID()
			if string(h[:]) == string(id) {
				return tx, nil
			}
		}
	}
	return nil, fmt.Errorf("tx not found")
}

// Balance returns SYNN token balance for an address.
func (s *LedgerService) Balance(addrHex string) (uint64, error) {
	addrHex = strings.TrimPrefix(addrHex, "0x")
	a, err := core.ParseAddress(addrHex)
	if err != nil {
		return 0, err
	}
	return s.ledger.BalanceOf(a), nil
}

// Info returns basic ledger information.
func (s *LedgerService) Info() map[string]interface{} {
	var height uint64
	var hash string
	if len(s.ledger.Blocks) > 0 {
		last := s.ledger.Blocks[len(s.ledger.Blocks)-1]
		height = last.Header.Height
		hash = last.Hash().Hex()
	}
	return map[string]interface{}{
		"height": height,
		"hash":   hash,
	}
}
