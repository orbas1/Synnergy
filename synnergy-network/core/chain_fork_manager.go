package core

import (
	"bytes"
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
func InitForkManager(l *Ledger) {
	forkMgr = &ChainForkManager{ledger: l, forks: make(map[string][]*Block)}
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
