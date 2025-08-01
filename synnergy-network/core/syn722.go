package core

import (
	"sync"
	"time"
)

type Mode int

const (
	Fungible Mode = iota
	NonFungible
)

type ModeChange struct {
	From Mode
	To   Mode
	Time time.Time
}

type SYN722Token struct {
	BaseToken
	mu       sync.RWMutex
	mode     Mode
	history  []ModeChange
	metadata map[Address]map[string]string
}

func NewSYN722Token(meta Metadata, init map[Address]uint64) (*SYN722Token, error) {
	t := &SYN722Token{mode: Fungible, metadata: make(map[Address]map[string]string)}
	tok, err := (Factory{}).Create(meta, init)
	if err != nil {
		return nil, err
	}
	t.BaseToken = *tok.(*BaseToken)
	return t, nil
}

func (t *SYN722Token) SetFungible() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.mode != Fungible {
		t.history = append(t.history, ModeChange{From: t.mode, To: Fungible, Time: time.Now().UTC()})
		t.mode = Fungible
	}
}

func (t *SYN722Token) SetNonFungible() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.mode != NonFungible {
		t.history = append(t.history, ModeChange{From: t.mode, To: NonFungible, Time: time.Now().UTC()})
		t.mode = NonFungible
	}
}

func (t *SYN722Token) Mode() Mode {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.mode
}

func (t *SYN722Token) ModeHistory() []ModeChange {
	t.mu.RLock()
	defer t.mu.RUnlock()
	hist := make([]ModeChange, len(t.history))
	copy(hist, t.history)
	return hist
}

func (t *SYN722Token) UpdateMetadata(addr Address, data map[string]string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.metadata == nil {
		t.metadata = make(map[Address]map[string]string)
	}
	cp := make(map[string]string, len(data))
	for k, v := range data {
		cp[k] = v
	}
	t.metadata[addr] = cp
}

func (t *SYN722Token) MetadataOf(addr Address) map[string]string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	m := t.metadata[addr]
	if m == nil {
		return nil
	}
	cp := make(map[string]string, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

// Opcode helpers used by the VM via opcode_dispatcher
func SYN722_SetFungible(id TokenID) bool {
	tok, ok := GetToken(id)
	if !ok {
		return false
	}
	s, ok := tok.(*SYN722Token)
	if !ok {
		return false
	}
	s.SetFungible()
	return true
}

func SYN722_SetNonFungible(id TokenID) bool {
	tok, ok := GetToken(id)
	if !ok {
		return false
	}
	s, ok := tok.(*SYN722Token)
	if !ok {
		return false
	}
	s.SetNonFungible()
	return true
}

func SYN722_Mode(id TokenID) Mode {
	tok, ok := GetToken(id)
	if !ok {
		return Fungible
	}
	s, ok := tok.(*SYN722Token)
	if !ok {
		return Fungible
	}
	return s.Mode()
}
