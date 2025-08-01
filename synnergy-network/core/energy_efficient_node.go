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

// EnergyNode_Start begins serving network traffic.
func EnergyNode_Start(en *EnergyEfficientNode) {
	if en == nil {
		return
	}
	go en.ListenAndServe()
}

// EnergyNode_Stop gracefully shuts down the node.
func EnergyNode_Stop(en *EnergyEfficientNode) error {
	if en == nil {
		return nil
	}
	return en.Close()
}

// EnergyNode_Record stores usage statistics for the validator.
func EnergyNode_Record(en *EnergyEfficientNode, txs uint64, kwh float64) error {
	if en == nil {
		return ErrInvalidNode
	}
	en.mu.Lock()
	defer en.mu.Unlock()
	return EnergyEff().RecordStats(en.validator, txs, kwh)
}

// EnergyNode_Efficiency returns the validator's current transactions-per-kWh.
func EnergyNode_Efficiency(en *EnergyEfficientNode) (float64, error) {
	if en == nil {
		return 0, ErrInvalidNode
	}
	en.mu.Lock()
	defer en.mu.Unlock()
	return EnergyEff().EfficiencyOf(en.validator)
}

// EnergyNode_NetworkAvg returns the network-wide transactions-per-kWh score.
func EnergyNode_NetworkAvg(en *EnergyEfficientNode) (float64, error) {
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
