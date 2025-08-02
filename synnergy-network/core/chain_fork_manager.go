package core

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// ForkInfo summarizes a fork branch.
type ForkInfo struct {
	Parent string `json:"parent"`
	Length int    `json:"length"`
}

// ChainForkManager tracks side branches and resolves simple forks.
type ChainForkManager struct {
	mu     sync.Mutex
	ledger *Ledger
	forks  map[string][]*Block
}

var forkMgr *ChainForkManager

// InitForkManager initialises the global fork manager with a ledger instance.
// It returns an error if the ledger is nil to avoid panics during block
// processing.
func InitForkManager(l *Ledger) error {
	if l == nil {
		return fmt.Errorf("nil ledger")
	}
	forkMgr = &ChainForkManager{ledger: l, forks: make(map[string][]*Block)}
	return nil
}

// AddForkBlock adds a new block to the ledger or stores it as a fork.
func AddForkBlock(b *Block) error {
	if forkMgr == nil {
		return fmt.Errorf("fork manager not initialised")
	}
	forkMgr.mu.Lock()
	defer forkMgr.mu.Unlock()

	var tipHash Hash
	if n := len(forkMgr.ledger.Blocks); n > 0 {
		tipHash = forkMgr.ledger.Blocks[n-1].Hash()
	}
	if bytes.Equal(b.Header.PrevHash, tipHash[:]) {
		return forkMgr.ledger.AddBlock(b)
	}

	parentHex := fmt.Sprintf("%x", b.Header.PrevHash)
	forkMgr.forks[parentHex] = append(forkMgr.forks[parentHex], b)
	logrus.WithFields(logrus.Fields{
		"parent": parentHex,
		"height": b.Header.Height,
	}).Info("block added to fork")
	return nil
}

// ListForks returns information about known forks.
func ListForks() []ForkInfo {
	if forkMgr == nil {
		return nil
	}
	forkMgr.mu.Lock()
	defer forkMgr.mu.Unlock()
	infos := make([]ForkInfo, 0, len(forkMgr.forks))
	for p, blocks := range forkMgr.forks {
		infos = append(infos, ForkInfo{Parent: p, Length: len(blocks)})
	}
	return infos
}

// ResolveForks appends fork blocks that extend the current tip.
func ResolveForks() error {
	if forkMgr == nil {
		return fmt.Errorf("fork manager not initialised")
	}
	forkMgr.mu.Lock()
	defer forkMgr.mu.Unlock()

	if len(forkMgr.ledger.Blocks) == 0 {
		return nil
	}
	tip := forkMgr.ledger.Blocks[len(forkMgr.ledger.Blocks)-1]
	h := tip.Hash()
	tipHex := fmt.Sprintf("%x", h[:])
	blocks, ok := forkMgr.forks[tipHex]
	if !ok {
		return nil
	}
	for _, blk := range blocks {
		if err := forkMgr.ledger.AddBlock(blk); err != nil {
			return err
		}
	}
	delete(forkMgr.forks, tipHex)
	logrus.WithField("new_height", forkMgr.ledger.Blocks[len(forkMgr.ledger.Blocks)-1].Header.Height).
		Info("fork resolved into main chain")
	return nil
}

// RecoverLongestFork searches known forks and, if a branch results in a longer
// chain than the current canonical tip, rewinds the ledger to the fork point and
// rebuilds using the longer branch. This provides automatic recovery from chain
// splits.
func RecoverLongestFork() error {
	if forkMgr == nil {
		return fmt.Errorf("fork manager not initialised")
	}
	forkMgr.mu.Lock()
	defer forkMgr.mu.Unlock()

	mainLen := len(forkMgr.ledger.Blocks)
	bestParent := ""
	var bestFork []*Block
	bestLen := mainLen

	for parentHex, blocks := range forkMgr.forks {
		parentBytes, err := hex.DecodeString(parentHex)
		if err != nil {
			logrus.WithField("parent", parentHex).WithError(err).Warn("invalid fork parent hash")
			continue
		}
		var ph Hash
		copy(ph[:], parentBytes)
		parent, ok := forkMgr.ledger.blockIndex[ph]
		if !ok {
			logrus.WithField("parent", parentHex).Warn("fork parent not found in ledger")
			continue
		}
		candLen := int(parent.Header.Height) + len(blocks) + 1
		if candLen > bestLen {
			bestLen = candLen
			bestParent = parentHex
			bestFork = blocks
		}
	}

	if bestParent == "" {
		return nil // no longer fork detected
	}

	parentBytes, _ := hex.DecodeString(bestParent)
	var parentHash Hash
	copy(parentHash[:], parentBytes)
	parentBlock := forkMgr.ledger.blockIndex[parentHash]
	parentHeight := int(parentBlock.Header.Height)

	newChain := append([]*Block{}, forkMgr.ledger.Blocks[:parentHeight+1]...)
	newChain = append(newChain, bestFork...)
	if err := forkMgr.ledger.RebuildChain(newChain); err != nil {
		return err
	}
	delete(forkMgr.forks, bestParent)
	logrus.WithField("new_height", forkMgr.ledger.Blocks[len(forkMgr.ledger.Blocks)-1].Header.Height).
		Info("chain reorganized to longest fork")
	return nil
}
