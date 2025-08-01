package core

import (
	"context"
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

// ------------------------------------------------------------------
// Simple flat gas calculator used by CLI stubs
// ------------------------------------------------------------------

type FlatGasCalculator struct{ Price uint64 }

func NewFlatGasCalculator(p uint64) *FlatGasCalculator { return &FlatGasCalculator{Price: p} }

func (f *FlatGasCalculator) Estimate(_ []byte) (uint64, error)     { return 0, nil }
func (f *FlatGasCalculator) Calculate(_ string, amt uint64) uint64 { return f.Price * amt }
