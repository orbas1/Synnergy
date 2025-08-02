package core

import (
	"context"
	"fmt"

	"github.com/wasmerio/wasmer-go/wasmer"
	Nodes "synnergy-network/core/Nodes"
)

// SuperNode provides enhanced network capabilities combining networking,
// ledger services and contract execution.
type SuperNode struct {
	*BaseNode
	ledger *Ledger
	ctx    context.Context
	cancel context.CancelFunc
}

// NewSuperNode initialises networking and a ledger for a multi purpose node.
func NewSuperNode(netCfg Config, ledCfg LedgerConfig) (*SuperNode, error) {
	ctx, cancel := context.WithCancel(context.Background())
	n, err := NewNode(netCfg)
	if err != nil {
		cancel()
		return nil, err
	}
	led, err := NewLedger(ledCfg)
	if err != nil {
		cancel()
		_ = n.Close()
		return nil, err
	}
	base := NewBaseNode(&NodeAdapter{n})
	return &SuperNode{BaseNode: base, ledger: led, ctx: ctx, cancel: cancel}, nil
}

// Close shuts down the node and ledger.
func (s *SuperNode) Close() error {
	s.cancel()
	return s.BaseNode.Close()
}

// ExecuteContract runs bytecode using the ledger's in-memory state.
func (s *SuperNode) ExecuteContract(code []byte) error {
	state, _ := NewInMemory()
	vm := NewHeavyVM(state, NewGasMeter(0), wasmer.NewEngine())
	_, err := vm.Execute(code, &VMContext{GasMeter: NewGasMeter(0), State: state})
	return err
}

// StoreData records arbitrary bytes under key in the ledger state.
func (s *SuperNode) StoreData(key string, data []byte) error {
	s.ledger.mu.Lock()
	defer s.ledger.mu.Unlock()
	s.ledger.State[key] = data
	return nil
}

// RetrieveData fetches bytes from the ledger state.
func (s *SuperNode) RetrieveData(key string) ([]byte, error) {
	s.ledger.mu.RLock()
	defer s.ledger.mu.RUnlock()
	val, ok := s.ledger.State[key]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return val, nil
}

// Ensure SuperNode implements the interface.
var _ Nodes.SuperNodeInterface = (*SuperNode)(nil)
