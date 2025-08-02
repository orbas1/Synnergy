package core

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// ExperimentalNode provides an isolated environment for testing new
// blockchain features without impacting the main network.
type ExperimentalNode struct {
	*NodeAdapter
	ledger   *Ledger
	features map[string][]byte
	mu       sync.Mutex
	logger   *logrus.Logger
}

// NewExperimentalNode creates a node and ledger for experimental use.
func NewExperimentalNode(netCfg Config, ledCfg LedgerConfig) (*ExperimentalNode, error) {
	n, err := NewNode(netCfg)
	if err != nil {
		return nil, err
	}
	led, err := NewLedger(ledCfg)
	if err != nil {
		_ = n.Close()
		return nil, err
	}
	return &ExperimentalNode{
		NodeAdapter: &NodeAdapter{n},
		ledger:      led,
		features:    make(map[string][]byte),
		logger:      logrus.StandardLogger(),
	}, nil
}

// StartTesting begins the node's network services.
func (e *ExperimentalNode) StartTesting() {
	go e.ListenAndServe()
	e.logger.Println("experimental node started")
}

// StopTesting stops the node and closes resources.
func (e *ExperimentalNode) StopTesting() error {
	e.logger.Println("experimental node stopping")
	return e.Close()
}

// DeployFeature stores experimental feature code on the ledger for testing.
func (e *ExperimentalNode) DeployFeature(name string, code []byte) error {
	if e == nil || e.ledger == nil {
		return fmt.Errorf("node not initialised")
	}
	e.mu.Lock()
	e.features[name] = code
	e.mu.Unlock()
	key := []byte("exp:" + name)
	return e.ledger.SetState(key, code)
}

// RollbackFeature removes a previously deployed feature.
func (e *ExperimentalNode) RollbackFeature(name string) error {
	if e == nil || e.ledger == nil {
		return fmt.Errorf("node not initialised")
	}
	e.mu.Lock()
	delete(e.features, name)
	e.mu.Unlock()
	key := []byte("exp:" + name)
	return e.ledger.DeleteState(key)
}

// SimulateTransaction adds a raw transaction to the ledger's pool.
func (e *ExperimentalNode) SimulateTransaction(data []byte) error {
	tx, err := DecodeTransaction(data)
	if err != nil {
		return err
	}
	e.ledger.AddToPool(tx)
	return nil
}

// TestContract deploys a contract into the experimental ledger.
func (e *ExperimentalNode) TestContract(bytecode []byte) error {
	addr := RandomAddress()
	c := Contract{Address: addr, Bytecode: bytecode}
	e.ledger.mu.Lock()
	e.ledger.Contracts[fmt.Sprintf("%x", addr)] = c
	e.ledger.mu.Unlock()
	return nil
}

// RandomAddress returns a pseudo-random address useful for testing.
func RandomAddress() Address {
	var a Address
	rand.Read(a[:])
	return a
}
