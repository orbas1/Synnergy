package core

import (
	"sync"
	"testing"
)

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
