package core

import (
	"context"
	"fmt"
	"sync"
)

var (
	ledgerOnce   sync.Once
	globalLedger *Ledger
	authSetOnce  sync.Once
	globalAuth   *AuthoritySet

	distOnce   sync.Once
	globalDist *TxDistributor
)

// InitLedger initialises the global ledger using OpenLedger at the given path.
func InitLedger(path string) error {
	var err error
	ledgerOnce.Do(func() {
		globalLedger, err = OpenLedger(path)
	})
	return err
}

// CurrentLedger returns the global ledger instance if initialised.
func CurrentLedger() *Ledger { return globalLedger }

// InitAuthoritySet stores a global authority set for CLI helpers.
func InitAuthoritySet(set *AuthoritySet) {
	authSetOnce.Do(func() { globalAuth = set })
}

// CurrentAuthoritySet returns the global authority set if initialised.
func CurrentAuthoritySet() *AuthoritySet { return globalAuth }

// InitTxDistributor initialises the global fee distributor.
func InitTxDistributor(l *Ledger) {
	distOnce.Do(func() { globalDist = NewTxDistributor(l) })
}

// CurrentTxDistributor returns the fee distributor if initialised.
func CurrentTxDistributor() *TxDistributor { return globalDist }

// ------------------------------------------------------------------
// TF gRPC stub client for AI module wiring
// ------------------------------------------------------------------

type tfStubClient struct{}

func NewTFStubClient(_ string) AIStubClient { return &tfStubClient{} }

func (tfStubClient) Anomaly(_ context.Context, _ *TFRequest) (*TFResponse, error) {
	return &TFResponse{}, nil
}
func (tfStubClient) FeeOpt(_ context.Context, _ *TFRequest) (*TFResponse, error) {
	return &TFResponse{}, nil
}
func (tfStubClient) Volume(_ context.Context, _ *TFRequest) (*TFResponse, error) {
	return &TFResponse{}, nil
}
func (tfStubClient) Inference(_ context.Context, _ *TFRequest) (*TFResponse, error) {
	return &TFResponse{}, nil
}
func (tfStubClient) Analyse(_ context.Context, _ *TFRequest) (*TFResponse, error) {
	return &TFResponse{}, nil
}

// ------------------------------------------------------------------
// Simple flat gas calculator used by CLI stubs
// ------------------------------------------------------------------

type FlatGasCalculator struct{ Price uint64 }

func NewFlatGasCalculator(p uint64) *FlatGasCalculator { return &FlatGasCalculator{Price: p} }

func (f *FlatGasCalculator) Estimate(_ []byte) (uint64, error)     { return 0, nil }
func (f *FlatGasCalculator) Calculate(_ string, amt uint64) uint64 { return f.Price * amt }

var (
	firewallOnce   sync.Once
	globalFirewall *Firewall
)

// InitFirewall initialises the global firewall instance.
func InitFirewall() {
	firewallOnce.Do(func() { globalFirewall = NewFirewall() })
}

// CurrentFirewall returns the global firewall if initialised.
func CurrentFirewall() *Firewall { return globalFirewall }
// ------------------------------------------------------------------
// DynamicGasCalculator parses bytecode and sums real gas costs
// ------------------------------------------------------------------

// DynamicGasCalculator implements GasCalculator using the opcode gas table. It
// estimates gas consumption by decoding 3-byte opcodes from a payload. Each
// opcode's base cost is retrieved via GasCost. If an unknown opcode is
// encountered the default gas cost is applied and an error is returned.
type DynamicGasCalculator struct{}

func NewDynamicGasCalculator() *DynamicGasCalculator { return &DynamicGasCalculator{} }

// Estimate walks the payload, treating it as a sequence of 3-byte opcodes. The
// total gas cost is the sum of GasCost for each opcode. An error is returned if
// the payload length is not a multiple of three or if an opcode fails to
// decode.
func (d *DynamicGasCalculator) Estimate(payload []byte) (uint64, error) {
	if len(payload)%3 != 0 {
		return 0, fmt.Errorf("invalid payload length %d", len(payload))
	}
	var total uint64
	for i := 0; i < len(payload); i += 3 {
		op, err := ParseOpcode(payload[i : i+3])
		if err != nil {
			return 0, err
		}
		total += GasCost(op)
	}
	return total, nil
}

// Calculate returns the gas for running the named opcode `amt` times. Unknown
// names fall back to DefaultGasCost.
func (d *DynamicGasCalculator) Calculate(name string, amt uint64) uint64 {
	if op, ok := nameToOp[name]; ok {
		return GasCost(op) * amt
	}
	return DefaultGasCost * amt
}
