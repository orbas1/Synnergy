package Tokens

import "testing"

// TestBalanceTableAddGet ensures Add and Get work as expected.
func TestBalanceTableAddGet(t *testing.T) {
	bt := NewBalanceTable()
	id := TokenID(1)
	var addr Address
	addr[0] = 0x01
	bt.Add(id, addr, 100)
	if got := bt.Get(id, addr); got != 100 {
		t.Fatalf("balance %d want 100", got)
	}
}

// TestBalanceTableSub verifies subtraction and error on insufficient funds.
func TestBalanceTableSub(t *testing.T) {
	bt := NewBalanceTable()
	id := TokenID(1)
	var addr Address
	addr[0] = 0x02
	bt.Add(id, addr, 50)
	if err := bt.Sub(id, addr, 30); err != nil {
		t.Fatalf("Sub returned error: %v", err)
	}
	if got := bt.Get(id, addr); got != 20 {
		t.Fatalf("balance %d want 20", got)
	}
	if err := bt.Sub(id, addr, 30); err == nil {
		t.Fatalf("expected error for insufficient balance")
	}
}

// TestBalanceTableSet ensures Set overwrites previous balances.
func TestBalanceTableSet(t *testing.T) {
	bt := NewBalanceTable()
	id := TokenID(1)
	var addr Address
	addr[0] = 0x03
	bt.Add(id, addr, 5)
	bt.Set(id, addr, 77)
	if got := bt.Get(id, addr); got != 77 {
		t.Fatalf("balance %d want 77", got)
	}
}
