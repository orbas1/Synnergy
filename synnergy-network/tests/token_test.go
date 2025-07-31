package core

import (
    "sync"
    "testing"
    "time"
)

//------------------------------------------------------------
// Minimal mocks for ledger + gas
//------------------------------------------------------------

type mockLedger struct{
    mu sync.Mutex
    gasCalls []uint64
    transfers []struct{from,to Address; amt uint64}
    approvals []struct{owner,spender Address; amt uint64}
}

func (m *mockLedger) EmitApproval(tok TokenID, owner, spender Address, amount uint64){
    m.mu.Lock(); defer m.mu.Unlock()
    m.approvals = append(m.approvals, struct{owner,spender Address; amt uint64}{owner,spender,amount})
}

func (m *mockLedger) EmitTransfer(tok TokenID, from, to Address, amount uint64){
    m.mu.Lock(); defer m.mu.Unlock()
    m.transfers = append(m.transfers, struct{from,to Address; amt uint64}{from,to,amount})
}

func (m *mockLedger) DeductGas(addr Address, amount uint64){
    m.mu.Lock(); defer m.mu.Unlock()
    m.gasCalls = append(m.gasCalls, amount)
}

func (m *mockLedger) WithinBlock(fn func() error) error { return fn() }

//------------------------------------------------------------
// Helpers
//------------------------------------------------------------

var addrA = Address{0x01}
var addrB = Address{0x02}

//------------------------------------------------------------
// Tests
//------------------------------------------------------------

func TestDeriveID(t *testing.T){
    got := deriveID(StdSYN20)
    want := TokenID(0x53000000 | uint32(StdSYN20)<<8)
    if got!=want { t.Fatalf("deriveID wrong: got %x want %x", got,want)}
}

func TestFactoryAndRegistry(t *testing.T){
    f := Factory{}
    meta := Metadata{"UnitTest","UT",0,StdSYN500,time.Time{},false,0}
    tok, err := f.Create(meta, map[Address]uint64{addrA: 100})
    if err!=nil { t.Fatalf("factory create err %v",err)}

    // verify registry
    got, ok := GetToken(tok.ID())
    if !ok || got.Meta().Symbol!="UT" {
        t.Fatalf("registry retrieval failed")
    }
    // supply tally
    if tok.Meta().TotalSupply!=100 {
        t.Errorf("total supply mismatch got %d", tok.Meta().TotalSupply)
    }
}

func prepareBaseToken(balanceA uint64) (*BaseToken,*mockLedger){
    bt := &BaseToken{
        id: deriveID(StdSYN600),
        meta: Metadata{Name:"Mock","MK",0,StdSYN600,time.Now(),false,0},
        balances: NewBalanceTable(),
        gas: Calculator{},
    }
    bt.balances.Set(bt.id, addrA, balanceA)
    l := &mockLedger{}
    bt.ledger = l
    return bt,l
}

func TestTransferSuccess(t *testing.T){
    tok, led := prepareBaseToken(1000)
    if err:=tok.Transfer(addrA, addrB, 400); err!=nil{
        t.Fatalf("transfer err %v",err)
    }
    if bal := tok.BalanceOf(addrA); bal!=600 {
        t.Errorf("sender balance %d want 600",bal)
    }
    if bal := tok.BalanceOf(addrB); bal!=400 {
        t.Errorf("recipient balance %d want 400",bal)
    }
    // gas deducted? expected 500+400/10000=500 (integer division)
    if len(led.gasCalls)!=1 || led.gasCalls[0]!=500 {
        t.Errorf("gas deduction wrong: %v", led.gasCalls)
    }
}

func TestTransferInsufficient(t *testing.T){
    tok, _ := prepareBaseToken(50)
    if err:=tok.Transfer(addrA, addrB, 100); err==nil{
        t.Fatalf("expected insufficient balance error")
    }
}

func TestApproveAndAllowance(t *testing.T){
    tok,_ := prepareBaseToken(0)
    if err:=tok.Approve(addrA, addrB, 999); err!=nil{
        t.Fatalf("approve err %v",err)
    }
    if all:=tok.Allowance(addrA, addrB); all!=999{
        t.Errorf("allowance %d want 999",all)
    }
}
