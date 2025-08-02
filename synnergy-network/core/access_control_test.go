package core

import (
	"strings"
	"sync"
	"testing"
)

type Address [20]byte

func (a Address) Hex() string {
	const hexdigits = "0123456789abcdef"
	out := make([]byte, 2+len(a)*2)
	copy(out, "0x")
	for i, v := range a {
		out[2+i*2] = hexdigits[v>>4]
		out[3+i*2] = hexdigits[v&0x0f]
	}
	return string(out)
}

type Ledger struct {
	State map[string][]byte
}

func (l *Ledger) HasState(k []byte) (bool, error) { _, ok := l.State[string(k)]; return ok, nil }
func (l *Ledger) SetState(k, v []byte) error      { l.State[string(k)] = v; return nil }
func (l *Ledger) DeleteState(k []byte) error      { delete(l.State, string(k)); return nil }

type iterator struct {
	keys [][]byte
	idx  int
}

func (l *Ledger) PrefixIterator(prefix []byte) *iterator {
	keys := make([][]byte, 0)
	for k := range l.State {
		if strings.HasPrefix(k, string(prefix)) {
			keys = append(keys, []byte(k))
		}
	}
	return &iterator{keys: keys}
}

func (it *iterator) Next() bool {
	if it.idx >= len(it.keys) {
		return false
	}
	it.idx++
	return true
}

func (it *iterator) Key() []byte  { return it.keys[it.idx-1] }
func (it *iterator) Error() error { return nil }

func TestAccessControllerCaching(t *testing.T) {
	led := &Ledger{State: make(map[string][]byte)}
	ac := NewAccessController(led)
	var addr Address
	addr[0] = 1
	role := "admin"
	if err := ac.GrantRole(addr, role); err != nil {
		t.Fatalf("grant: %v", err)
	}
	if !ac.HasRole(addr, role) {
		t.Fatalf("expected role present")
	}
	roles, err := ac.ListRoles(addr)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(roles) != 1 || roles[0] != role {
		t.Fatalf("unexpected roles %v", roles)
	}
	if err := ac.RevokeRole(addr, role); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if ac.HasRole(addr, role) {
		t.Fatalf("expected role removed")
	}
}

func BenchmarkAccessControllerHasRole(b *testing.B) {
	led := &Ledger{State: make(map[string][]byte)}
	ac := NewAccessController(led)
	var addr Address
	role := "bench"
	if err := ac.GrantRole(addr, role); err != nil {
		b.Fatalf("grant: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !ac.HasRole(addr, role) {
			b.Fatal("missing role")
		}
	}
}

func TestAccessControllerConcurrent(t *testing.T) {
	led := &Ledger{State: make(map[string][]byte)}
	ac := NewAccessController(led)
	var addr Address
	role := "worker"
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = ac.GrantRole(addr, role)
		}()
	}
	wg.Wait()
	if !ac.HasRole(addr, role) {
		t.Fatalf("expected role present")
	}
	roles, err := ac.ListRoles(addr)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(roles) != 1 || roles[0] != role {
		t.Fatalf("unexpected roles %v", roles)
	}
}
