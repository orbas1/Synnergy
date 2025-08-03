package core_test

import (
	"math/big"
	. "synnergy-network/core"
	"testing"
)

// TestRegisterBridge ensures a bridge is persisted and retrievable
func TestRegisterBridge(t *testing.T) {
	st := NewInMemoryStore()
	SetStore(st)
	SetBroadcaster(func(string, []byte) error { return nil })

	b := Bridge{SourceChain: "src", TargetChain: "dst", Relayer: Address{0x01}}
	if err := RegisterBridge(b); err != nil {
		t.Fatalf("register err %v", err)
	}

	bridges, err := ListBridges()
	if err != nil || len(bridges) != 1 {
		t.Fatalf("list got %v err %v", bridges, err)
	}
	if bridges[0].SourceChain != "src" {
		t.Fatalf("unexpected bridge data")
	}

	out, err := GetBridge(bridges[0].ID)
	if err != nil || out.TargetChain != "dst" {
		t.Fatalf("get err %v data %+v", err, out)
	}
}

// simple ledger mock implementing only required methods

type simpleLedger struct {
	transfers []string
	mints     []string
	burns     []string
}

func (l *simpleLedger) Transfer(from, to Address, amount uint64) error {
	l.transfers = append(l.transfers, from.String()+"->"+to.String())
	return nil
}
func (l *simpleLedger) Mint(addr Address, amount uint64) error {
	l.mints = append(l.mints, addr.String())
	return nil
}
func (l *simpleLedger) Burn(addr Address, amount uint64) error {
	l.burns = append(l.burns, addr.String())
	return nil
}

// minimal Context

type testCtx struct {
	Caller Address
	State  *simpleLedger
}

func (c *testCtx) Gas(uint64) error  { return nil }
func (c *testCtx) Call(string) error { return nil }

// Wrap into core.Context struct for Transfer/Mint helpers

func (c *testCtx) toCoreCtx() *Context {
	return &Context{Caller: c.Caller, State: c}
}

// implement StateRW subset
func (l *simpleLedger) GetState([]byte) ([]byte, error)                     { return nil, nil }
func (l *simpleLedger) SetState([]byte, []byte) error                       { return nil }
func (l *simpleLedger) DeleteState([]byte) error                            { return nil }
func (l *simpleLedger) HasState([]byte) (bool, error)                       { return false, nil }
func (l *simpleLedger) PrefixIterator([]byte) StateIterator                 { return nil }
func (l *simpleLedger) IsIDTokenHolder(Address) bool                        { return false }
func (l *simpleLedger) Snapshot(func() error) error                         { return nil }
func (l *simpleLedger) MintLP(Address, PoolID, uint64) error                { return nil }
func (l *simpleLedger) MintToken(Address, uint64) error                     { return nil }
func (l *simpleLedger) TransferState(from, to Address, amount uint64) error { return nil }
func (l *simpleLedger) BalanceOf(Address) uint64                            { return 0 }
func (l *simpleLedger) NonceOf(Address) uint64                              { return 0 }
func (l *simpleLedger) BurnLP(Address, PoolID, uint64) error                { return nil }
func (l *simpleLedger) Get([]byte, []byte) ([]byte, error)                  { return nil, nil }
func (l *simpleLedger) Set([]byte, []byte, []byte) error                    { return nil }
func (l *simpleLedger) GetCode(Address) []byte                              { return nil }
func (l *simpleLedger) GetCodeHash(Address) Hash                            { return Hash{} }
func (l *simpleLedger) AddLog(*Log)                                         {}
func (l *simpleLedger) CreateContract(Address, []byte, *big.Int, uint64) (Address, []byte, bool, error) {
	return AddressZero, nil, false, nil
}
func (l *simpleLedger) DelegateCall(Address, Address, []byte, *big.Int, uint64) error { return nil }
func (l *simpleLedger) Call(Address, Address, []byte, *big.Int, uint64) ([]byte, error) {
	return nil, nil
}
func (l *simpleLedger) GetContract(Address) (*Contract, error)           { return nil, nil }
func (l *simpleLedger) GetToken(TokenID) (Token, error)                  { return nil, nil }
func (l *simpleLedger) GetTokenBalance(Address, TokenID) (uint64, error) { return 0, nil }
func (l *simpleLedger) SetTokenBalance(Address, TokenID, uint64) error   { return nil }
func (l *simpleLedger) GetTokenSupply(TokenID) (uint64, error)           { return 0, nil }
func (l *simpleLedger) CallCode(Address, Address, []byte, *big.Int, uint64) ([]byte, bool, error) {
	return nil, false, nil
}
func (l *simpleLedger) CallContract(Address, Address, []byte, *big.Int, uint64) ([]byte, bool, error) {
	return nil, false, nil
}
func (l *simpleLedger) StaticCall(Address, Address, []byte, uint64) ([]byte, bool, error) {
	return nil, false, nil
}
func (l *simpleLedger) SelfDestruct(Address, Address) {}
