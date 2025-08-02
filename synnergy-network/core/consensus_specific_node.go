package core

import (
	"context"
	"encoding/hex"
	"github.com/sirupsen/logrus"
)

// ConsensusSpecificNode implements a node tuned for a particular consensus algorithm.
type ConsensusSpecificNode struct {
	*Node
	Ledger *Ledger
	engine *SynnergyConsensus
	logger *logrus.Logger
}

// NewConsensusSpecificNode constructs a consensus node with the provided engine and ledger.
func NewConsensusSpecificNode(cfg Config, led *Ledger, engine *SynnergyConsensus, lg *logrus.Logger) (*ConsensusSpecificNode, error) {
	n, err := NewNode(cfg)
	if err != nil {
		return nil, err
	}
	if lg == nil {
		lg = logrus.New()
	}
	return &ConsensusSpecificNode{Node: n, Ledger: led, engine: engine, logger: lg}, nil
}

// StartConsensus boots the consensus engine in a new goroutine.
func (csn *ConsensusSpecificNode) StartConsensus() error {
	if csn.engine == nil {
		return nil
	}
	go csn.engine.Start(context.Background())
	csn.logger.Info("consensus engine started")
	return nil
}

// StopConsensus stops the consensus engine and network listener.
func (csn *ConsensusSpecificNode) StopConsensus() error {
	if csn.engine != nil {
		csn.engine.Stop()
	}
	return csn.Close()
}

// SubmitBlock appends a block to the underlying ledger.
func (csn *ConsensusSpecificNode) SubmitBlock(b *Block) error {
	if csn.Ledger == nil {
		return nil
	}
	return csn.Ledger.AppendBlock(b)
}

// ProcessTx adds a transaction to the ledger pool for future blocks.
func (csn *ConsensusSpecificNode) ProcessTx(tx *Transaction) error {
	if csn.Ledger == nil || tx == nil {
		return nil
	}
	csn.Ledger.mu.Lock()
	defer csn.Ledger.mu.Unlock()
	if csn.Ledger.TxPool == nil {
		csn.Ledger.TxPool = make(map[string]*Transaction)
	}
	csn.Ledger.TxPool[hex.EncodeToString(tx.Hash[:])] = tx
	return nil
}
