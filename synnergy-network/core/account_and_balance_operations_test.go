package core

import "testing"

// Minimal versions of core types required to exercise the account manager logic
// in isolation. The full implementations live elsewhere in the codebase, but
// for unit testing we only need an address type capable of being used as a map
// key and a ledger structure holding token balances.
//
// Defining them here keeps the tests self‑contained and avoids pulling in the
// entire core package dependency graph, which currently contains unrelated
// components that do not compile.

// Address is a 20‑byte identifier. Only the String method used by the account
// manager is implemented.
type Address [20]byte

func (a Address) String() string { return string(a[:]) }

// Ledger holds token balances for addresses. Additional fields present in the
// production ledger are omitted as they are unnecessary for these tests.
type Ledger struct {
	TokenBalances map[string]uint64
}

func TestAccountManagerCreateAndBalance(t *testing.T) {
	ledger := &Ledger{TokenBalances: make(map[string]uint64)}
	am := NewAccountManager(ledger)
	var addr Address
	copy(addr[:], []byte("address-1-000000"))

	if err := am.CreateAccount(addr); err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	bal, err := am.Balance(addr)
	if err != nil {
		t.Fatalf("Balance returned error: %v", err)
	}
	if bal != 0 {
		t.Fatalf("expected balance 0, got %d", bal)
	}

	if err := am.CreateAccount(addr); err == nil {
		t.Fatalf("expected error when creating existing account")
	}
}

func TestAccountManagerTransferAndDelete(t *testing.T) {
	ledger := &Ledger{TokenBalances: make(map[string]uint64)}
	am := NewAccountManager(ledger)

	var src, dst Address
	copy(src[:], []byte("source-address-000"))
	copy(dst[:], []byte("dest-address-00000"))

	if err := am.CreateAccount(src); err != nil {
		t.Fatalf("CreateAccount src failed: %v", err)
	}
	ledger.TokenBalances[src.String()] = 100
	if err := am.CreateAccount(dst); err != nil {
		t.Fatalf("CreateAccount dst failed: %v", err)
	}

	if err := am.Transfer(src, dst, 40); err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}
	if ledger.TokenBalances[src.String()] != 60 {
		t.Fatalf("src expected 60, got %d", ledger.TokenBalances[src.String()])
	}
	if ledger.TokenBalances[dst.String()] != 40 {
		t.Fatalf("dst expected 40, got %d", ledger.TokenBalances[dst.String()])
	}

	if err := am.DeleteAccount(src); err != nil {
		t.Fatalf("DeleteAccount failed: %v", err)
	}
	if _, ok := ledger.TokenBalances[src.String()]; ok {
		t.Fatalf("source account still exists after deletion")
	}
}
