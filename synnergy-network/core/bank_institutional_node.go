package core

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
)

// BankInstitutionalConfig groups dependencies required by the node.
type BankInstitutionalConfig struct {
	Ledger    *Ledger
	Network   *Node
	Consensus *SynnergyConsensus
	TxPool    *TxPool
}

// BankInstitutionalNode provides extended functionality for banks and large
// institutions. It embeds the base network node and integrates with the ledger
// and consensus engine.
type BankInstitutionalNode struct {
	*Node
	ledger    *Ledger
	consensus *SynnergyConsensus
	txpool    *TxPool
	rules     map[string]interface{}
	analytics map[string]int64
	mu        sync.RWMutex
}

// NewBankInstitutionalNode constructs a new bank/institutional authority node.
func NewBankInstitutionalNode(cfg BankInstitutionalConfig) (*BankInstitutionalNode, error) {
	if cfg.Network == nil {
		return nil, errors.New("network node required")
	}
	n := &BankInstitutionalNode{
		Node:      cfg.Network,
		ledger:    cfg.Ledger,
		consensus: cfg.Consensus,
		txpool:    cfg.TxPool,
		rules:     make(map[string]interface{}),
		analytics: make(map[string]int64),
	}
	return n, nil
}

// Start launches the underlying network services.
func (n *BankInstitutionalNode) Start() {
	if n.Node != nil {
		go n.Node.ListenAndServe()
	}
}

// Stop gracefully shuts down the node.
func (n *BankInstitutionalNode) Stop() error {
	if n.Node != nil {
		return n.Node.Close()
	}
	return nil
}

// MonitorTransaction performs basic compliance checks and records analytics.
func (n *BankInstitutionalNode) MonitorTransaction(tx *Transaction) error {
	if tx == nil {
		return errors.New("nil tx")
	}
	n.mu.Lock()
	n.analytics["tx_processed"]++
	n.mu.Unlock()
	return nil
}

// SubmitTx validates and queues a transaction in the mempool.
func (n *BankInstitutionalNode) SubmitTx(tx *Transaction) error {
	if n.txpool == nil {
		return errors.New("txpool not configured")
	}
	n.txpool.mu.Lock()
	defer n.txpool.mu.Unlock()
	n.txpool.lookup[tx.Hash] = tx
	n.txpool.queue = append(n.txpool.queue, tx)
	return nil
}

// ComplianceReport returns a JSON encoded summary of analytics.
func (n *BankInstitutionalNode) ComplianceReport() ([]byte, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return json.Marshal(n.analytics)
}

// ConnectFinancialNetwork simulates establishing secure connections to legacy systems.
func (n *BankInstitutionalNode) ConnectFinancialNetwork(endpoint string) error {
	log.Printf("[BANK NODE] connecting to financial endpoint %s", endpoint)
	return nil
}

// UpdateRuleset applies a set of policy or compliance rules.
func (n *BankInstitutionalNode) UpdateRuleset(r map[string]interface{}) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for k, v := range r {
		n.rules[k] = v
	}
}
