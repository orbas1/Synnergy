package core

import (
	"sync"
	"time"
)

type PriceRecord struct {
	Time  time.Time
	Price uint64
}

type SYN1967Token struct {
	BaseToken
	Commodity     string
	Unit          string
	Certification string
	Trace         []string
	price         uint64
	history       []PriceRecord
	mu            sync.RWMutex
}

func NewSYN1967Token(meta Metadata, commodity, unit string, price uint64) *SYN1967Token {
	if meta.Standard == 0 {
		meta.Standard = StdSYN1967
	}
	t := &SYN1967Token{
		BaseToken: BaseToken{
			id:       deriveID(meta.Standard),
			meta:     meta,
			balances: NewBalanceTable(),
		},
		Commodity: commodity,
		Unit:      unit,
		price:     price,
	}
	t.history = append(t.history, PriceRecord{Time: time.Now().UTC(), Price: price})
	return t
}

func (t *SYN1967Token) UpdatePrice(p uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.price = p
	t.history = append(t.history, PriceRecord{Time: time.Now().UTC(), Price: p})
}

func (t *SYN1967Token) CurrentPrice() uint64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.price
}

func (t *SYN1967Token) PriceHistory() []PriceRecord {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]PriceRecord, len(t.history))
	copy(out, t.history)
	return out
}

func (t *SYN1967Token) AddCertification(cert string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Certification = cert
}

func (t *SYN1967Token) AddTrace(info string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Trace = append(t.Trace, info)
}
