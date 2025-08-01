package core

import (
	"sync"
	"time"
)

type ETFRecord struct {
	ETFID           string
	Name            string
	TotalShares     uint64
	AvailableShares uint64
	CurrentPrice    uint64
}

type ETFShare struct {
	Holder  Address
	Shares  int64
	Issued  time.Time
	Updated time.Time
}

type SYN3300Token struct {
	*BaseToken
	ETF          ETFRecord
	shareHistory map[Address][]ETFShare
	mu           sync.RWMutex
}

var syn3300Registry struct {
	sync.RWMutex
	tokens map[TokenID]*SYN3300Token
}

func init() {
	syn3300Registry.tokens = make(map[TokenID]*SYN3300Token)
}

func NewSYN3300(meta Metadata, etf ETFRecord, init map[Address]uint64) (*SYN3300Token, error) {
	tok, err := (Factory{}).Create(meta, init)
	if err != nil {
		return nil, err
	}
	s := &SYN3300Token{
		BaseToken:    tok.(*BaseToken),
		ETF:          etf,
		shareHistory: make(map[Address][]ETFShare),
	}
	syn3300Registry.Lock()
	syn3300Registry.tokens[s.ID()] = s
	syn3300Registry.Unlock()
	return s, nil
}

func GetSYN3300(id TokenID) (*SYN3300Token, bool) {
	syn3300Registry.RLock()
	defer syn3300Registry.RUnlock()
	t, ok := syn3300Registry.tokens[id]
	return t, ok
}

func ListSYN3300() []*SYN3300Token {
	syn3300Registry.RLock()
	defer syn3300Registry.RUnlock()
	out := make([]*SYN3300Token, 0, len(syn3300Registry.tokens))
	for _, t := range syn3300Registry.tokens {
		out = append(out, t)
	}
	return out
}

func (t *SYN3300Token) UpdatePrice(p uint64) {
	t.mu.Lock()
	t.ETF.CurrentPrice = p
	t.mu.Unlock()
}

func (t *SYN3300Token) FractionalMint(to Address, shares uint64) error {
	if err := t.Mint(to, shares); err != nil {
		return err
	}
	t.mu.Lock()
	t.ETF.TotalShares += shares
	t.ETF.AvailableShares += shares
	t.shareHistory[to] = append(t.shareHistory[to], ETFShare{Holder: to, Shares: int64(shares), Issued: time.Now(), Updated: time.Now()})
	t.mu.Unlock()
	return nil
}

func (t *SYN3300Token) FractionalBurn(from Address, shares uint64) error {
	if err := t.Burn(from, shares); err != nil {
		return err
	}
	t.mu.Lock()
	if t.ETF.TotalShares >= shares {
		t.ETF.TotalShares -= shares
	}
	if t.ETF.AvailableShares >= shares {
		t.ETF.AvailableShares -= shares
	}
	t.shareHistory[from] = append(t.shareHistory[from], ETFShare{Holder: from, Shares: -int64(shares), Issued: time.Now(), Updated: time.Now()})
	t.mu.Unlock()
	return nil
}

func (t *SYN3300Token) GetETFInfo() ETFRecord {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.ETF
}
