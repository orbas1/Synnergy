package core

import (
	"fmt"
	"sync"
)

// EnergyEfficientNode combines a network node with energy efficiency tracking.
// It records energy usage statistics for a validator and exposes convenience
// helpers to query efficiency metrics.
type EnergyEfficientNode struct {
	*Node
	ledger    *Ledger
	validator Address
	mu        sync.Mutex
}

// NewEnergyNode initialises a node and attaches the energy efficiency engine.
func NewEnergyNode(cfg Config, led *Ledger, validator Address) (*EnergyEfficientNode, error) {
	n, err := NewNode(cfg)
	if err != nil {
		return nil, err
	}
	if led == nil {
		return nil, ErrInvalidLedger
	}
	InitEnergyEfficiency(led)
	return &EnergyEfficientNode{Node: n, ledger: led, validator: validator}, nil
}

// EnergyNodeStart begins serving network traffic.
func EnergyNodeStart(en *EnergyEfficientNode) {
	if en == nil {
		return
	}
	go en.ListenAndServe()
}

// EnergyNodeStop gracefully shuts down the node.
func EnergyNodeStop(en *EnergyEfficientNode) error {
	if en == nil {
		return nil
	}
	return en.Close()
}

// EnergyNodeRecord stores usage statistics for the validator.
func EnergyNodeRecord(en *EnergyEfficientNode, txs uint64, kwh float64) error {
	if en == nil {
		return ErrInvalidNode
	}
	if kwh <= 0 {
		return fmt.Errorf("kwh must be positive")
	}
	en.mu.Lock()
	defer en.mu.Unlock()
	return EnergyEff().RecordStats(en.validator, txs, kwh)
}

// EnergyNodeEfficiency returns the validator's current transactions-per-kWh.
func EnergyNodeEfficiency(en *EnergyEfficientNode) (float64, error) {
	if en == nil {
		return 0, ErrInvalidNode
	}
	en.mu.Lock()
	defer en.mu.Unlock()
	return EnergyEff().EfficiencyOf(en.validator)
}

// EnergyNodeNetworkAvg returns the network-wide transactions-per-kWh score.
func EnergyNodeNetworkAvg(en *EnergyEfficientNode) (float64, error) {
	if en == nil {
		return 0, ErrInvalidNode
	}
	return EnergyEff().NetworkAverage()
}

var (
	// ErrInvalidLedger is returned when a nil ledger is supplied.
	ErrInvalidLedger = fmt.Errorf("invalid ledger")
	// ErrInvalidNode is returned when the node reference is nil.
	ErrInvalidNode = fmt.Errorf("invalid node")
)
