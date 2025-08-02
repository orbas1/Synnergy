package Tokens

import "testing"

// helper to create distinct addresses
func addr(b byte) Address {
	var a Address
	a[19] = b
	return a
}

func TestBaseTokenLifecycle(t *testing.T) {
	regMu.Lock()
	reg = make(map[TokenID]Token)
	regMu.Unlock()
	owner := addr(1)
	spender := addr(2)
	recipient := addr(3)

	meta := Metadata{Name: "TestToken", Symbol: "TT", Decimals: 2, Standard: StdSYN10}
	tokInt, err := (Factory{}).Create(meta, map[Address]uint64{owner: 100})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	tok := tokInt.(*BaseToken)

	if bal := tok.BalanceOf(owner); bal != 100 {
		t.Fatalf("owner balance %d", bal)
	}

	if err := tok.Transfer(owner, recipient, 40); err != nil {
		t.Fatalf("transfer: %v", err)
	}
	if bal := tok.BalanceOf(recipient); bal != 40 {
		t.Fatalf("recipient balance %d", bal)
	}

	if err := tok.Approve(owner, spender, 50); err != nil {
		t.Fatalf("approve: %v", err)
	}
	if err := tok.TransferFrom(owner, spender, recipient, 30); err != nil {
		t.Fatalf("transferFrom: %v", err)
	}
	if al := tok.Allowance(owner, spender); al != 20 {
		t.Fatalf("allowance %d", al)
	}

	if err := tok.Burn(recipient, 10); err != nil {
		t.Fatalf("burn: %v", err)
	}
	if tok.meta.TotalSupply != 90 {
		t.Fatalf("total supply %d", tok.meta.TotalSupply)
	}
}

func TestFixedSupplyMint(t *testing.T) {
	regMu.Lock()
	reg = make(map[TokenID]Token)
	regMu.Unlock()
	owner := addr(1)
	meta := Metadata{Name: "Fixed", Symbol: "FX", Standard: StdSYN20, FixedSupply: true}
	tok := &BaseToken{id: deriveID(meta.Standard), meta: meta, balances: NewBalanceTable()}
	if err := tok.Mint(owner, 1); err == nil {
		t.Fatalf("expected mint to fail for fixed supply")
	}
}
