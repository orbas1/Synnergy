package core

import (
	"fmt"
	"sync"
	"time"
)

// FuturesContract defines the essential metadata of a futures contract.
type FuturesContract struct {
	ContractID    string
	Underlying    string
	Expiration    time.Time
	ContractSize  uint64
	InitialMargin uint64
}

// FuturesPosition captures an investor's position in a futures token.
type FuturesPosition struct {
	Size       uint64
	EntryPrice uint64
	Long       bool
	Margin     uint64
	PnL        int64
}

// FuturesToken extends BaseToken with futures trading functionality.
type FuturesToken struct {
	*BaseToken
	Contract  FuturesContract
	Positions map[Address]*FuturesPosition
	Price     uint64
	mu        sync.RWMutex
}

// NewFuturesToken creates a futures token with the provided contract metadata.
func NewFuturesToken(meta Metadata, contract FuturesContract, init map[Address]uint64) (*FuturesToken, error) {
	if meta.Standard == 0 {
		meta.Standard = StdSYN3600
	}
	if meta.Created.IsZero() {
		meta.Created = time.Now().UTC()
	}
	ft := &FuturesToken{
		BaseToken: &BaseToken{id: deriveID(meta.Standard), meta: meta, balances: NewBalanceTable()},
		Contract:  contract,
		Positions: make(map[Address]*FuturesPosition),
	}
	for a, v := range init {
		ft.balances.Set(ft.id, a, v)
		ft.meta.TotalSupply += v
	}
	RegisterToken(ft.BaseToken)
	return ft, nil
}

// UpdatePrice sets the current settlement price for the futures contract.
func (ft *FuturesToken) UpdatePrice(price uint64) {
	ft.mu.Lock()
	ft.Price = price
	ft.mu.Unlock()
}

// OpenPosition records a new speculative position for an investor.
func (ft *FuturesToken) OpenPosition(addr Address, size, entryPrice uint64, long bool, margin uint64) error {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	if margin < ft.Contract.InitialMargin {
		return fmt.Errorf("insufficient margin")
	}
	ft.Positions[addr] = &FuturesPosition{Size: size, EntryPrice: entryPrice, Long: long, Margin: margin}
	return nil
}

// ClosePosition settles an existing position and returns the profit or loss.
func (ft *FuturesToken) ClosePosition(addr Address, exitPrice uint64) (int64, error) {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	pos, ok := ft.Positions[addr]
	if !ok {
		return 0, fmt.Errorf("position not found")
	}
	var pnl int64
	if pos.Long {
		pnl = int64(exitPrice-pos.EntryPrice) * int64(pos.Size)
	} else {
		pnl = int64(pos.EntryPrice-exitPrice) * int64(pos.Size)
	}
	delete(ft.Positions, addr)
	pos.PnL = pnl
	return pnl, nil
}
