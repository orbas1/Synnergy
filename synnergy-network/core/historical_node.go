package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// HistoricalNode maintains a complete archive of all blocks and exposes
// retrieval helpers for audits and analytics.
type HistoricalNode struct {
	*BaseNode
	ledger     *Ledger
	archiveDir string
	mu         sync.RWMutex
}

// NewHistoricalNode creates a networking node wired to a ledger. Blocks are
// written to archiveDir in JSON format for long term storage.
func NewHistoricalNode(cfg Config, led *Ledger, archiveDir string) (*HistoricalNode, error) {
	n, err := NewNode(cfg)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		return nil, err
	}
	base := NewBaseNode(&NodeAdapter{n})
	return &HistoricalNode{
		BaseNode:   base,
		ledger:     led,
		archiveDir: archiveDir,
	}, nil
}

// SyncFromLedger persists every block currently stored in the ledger.
func (h *HistoricalNode) SyncFromLedger() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, blk := range h.ledger.Blocks {
		if err := h.ArchiveBlock(blk); err != nil {
			return err
		}
	}
	return nil
}

// ArchiveBlock writes a block to disk using its height as filename.
func (h *HistoricalNode) ArchiveBlock(b *Block) error {
	data, err := json.Marshal(b)
	if err != nil {
		return err
	}
	path := filepath.Join(h.archiveDir, fmt.Sprintf("%d.json", b.Header.Height))
	return os.WriteFile(path, data, 0o644)
}

// BlockByHeight loads a block from the archive directory.
func (h *HistoricalNode) BlockByHeight(height uint64) (*Block, error) {
	path := filepath.Join(h.archiveDir, fmt.Sprintf("%d.json", height))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var blk Block
	if err := json.Unmarshal(data, &blk); err != nil {
		return nil, err
	}
	return &blk, nil
}

// RangeBlocks returns blocks in [start,end] inclusive.
func (h *HistoricalNode) RangeBlocks(start, end uint64) ([]*Block, error) {
	var out []*Block
	for i := start; i <= end; i++ {
		blk, err := h.BlockByHeight(i)
		if err != nil {
			return nil, err
		}
		out = append(out, blk)
	}
	return out, nil
}
