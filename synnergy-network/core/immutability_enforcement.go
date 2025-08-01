package core

import (
	"bytes"
	"fmt"
	"sync"
)

// ImmutabilityEnforcer ensures the genesis block cannot be altered.
type ImmutabilityEnforcer struct {
	ledger       *Ledger
	genesisBlock *Block
	genesisHash  [32]byte
	mu           sync.RWMutex
}

var (
	immOnce        sync.Once
	globalEnforcer *ImmutabilityEnforcer
)

// NewImmutabilityEnforcer configures the enforcer for a given ledger.
func NewImmutabilityEnforcer(l *Ledger) (*ImmutabilityEnforcer, error) {
	if l == nil {
		return nil, fmt.Errorf("immutability: nil ledger")
	}
	if len(l.Blocks) == 0 {
		return nil, fmt.Errorf("immutability: no genesis block")
	}
	g := l.Blocks[0]
	return &ImmutabilityEnforcer{
		ledger:       l,
		genesisBlock: g,
		genesisHash:  g.Hash(),
	}, nil
}

// InitImmutability initialises a global enforcer for CLI helpers.
func InitImmutability(l *Ledger) error {
	var err error
	immOnce.Do(func() {
		globalEnforcer, err = NewImmutabilityEnforcer(l)
	})
	return err
}

// CurrentEnforcer returns the global enforcer if initialised.
func CurrentEnforcer() *ImmutabilityEnforcer { return globalEnforcer }

// VerifyChain ensures the ledger's chain links are intact.
func (ie *ImmutabilityEnforcer) VerifyChain() error {
	ie.mu.RLock()
	defer ie.mu.RUnlock()

	if len(ie.ledger.Blocks) == 0 {
		return fmt.Errorf("immutability: empty ledger")
	}
	if ie.genesisHash != ie.ledger.Blocks[0].Hash() {
		return fmt.Errorf("immutability: genesis block modified")
	}
	for i := 1; i < len(ie.ledger.Blocks); i++ {
		prev := ie.ledger.Blocks[i-1].Hash()
		if !bytes.Equal(ie.ledger.Blocks[i].Header.PrevHash, prev[:]) {
			return fmt.Errorf("immutability: invalid prev hash at height %d", i)
		}
	}
	return nil
}

// RestoreChain resets the genesis block if it was altered.
func (ie *ImmutabilityEnforcer) RestoreChain() error {
	ie.mu.Lock()
	defer ie.mu.Unlock()

	if len(ie.ledger.Blocks) == 0 {
		return fmt.Errorf("immutability: empty ledger")
	}
	if ie.ledger.Blocks[0].Hash() != ie.genesisHash {
		ie.ledger.Blocks[0] = ie.genesisBlock
	}
	return nil
}
