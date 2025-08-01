package core

import (
	"sync"
	"time"
)

type SwapRecord struct {
	PartnerChain string
	From         Address
	To           Address
	Amount       uint64
	Completed    bool
	Created      time.Time
}

type SYN1200Token struct {
	*BaseToken
	mu      sync.Mutex
	Bridges map[string]Address
	Swaps   map[string]*SwapRecord
}

func NewSYN1200(meta Metadata, init map[Address]uint64) (*SYN1200Token, error) {
	tok, err := (Factory{}).Create(meta, init)
	if err != nil {
		return nil, err
	}
	return &SYN1200Token{
		BaseToken: tok.(*BaseToken),
		Bridges:   make(map[string]Address),
		Swaps:     make(map[string]*SwapRecord),
	}, nil
}

func (t *SYN1200Token) AddBridge(chain string, addr Address) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Bridges == nil {
		t.Bridges = make(map[string]Address)
	}
	t.Bridges[chain] = addr
}

func (t *SYN1200Token) AtomicSwap(id, partnerChain string, from, to Address, amount uint64) error {
	if err := t.Transfer(from, to, amount); err != nil {
		return err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Swaps == nil {
		t.Swaps = make(map[string]*SwapRecord)
	}
	t.Swaps[id] = &SwapRecord{
		PartnerChain: partnerChain,
		From:         from,
		To:           to,
		Amount:       amount,
		Created:      time.Now().UTC(),
	}
	return nil
}

func (t *SYN1200Token) CompleteSwap(id string) {
	t.mu.Lock()
	if rec, ok := t.Swaps[id]; ok {
		rec.Completed = true
	}
	t.mu.Unlock()
}

func (t *SYN1200Token) GetSwap(id string) (*SwapRecord, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	rec, ok := t.Swaps[id]
	return rec, ok
}
