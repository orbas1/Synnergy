package core

import (
    "encoding/json"
    "errors"
    "testing"
    "time"
)

// ------------------------------------------------------------
// Mock implementations
// ------------------------------------------------------------

type mockLedger struct {
    state    map[string][]byte
    transfers []string // record "from->to:amt"
    balances  map[string]uint64
}

func newMockLedger() *mockLedger {
    return &mockLedger{state: make(map[string][]byte), balances: make(map[string]uint64)}
}

func (m *mockLedger) SetState(k, v []byte) {
    m.state[string(k)] = v
}

func (m *mockLedger) GetState(k []byte) ([]byte, error) {
    if v, ok := m.state[string(k)]; ok {
        return v, nil
    }
    return nil, nil
}

func (m *mockLedger) HasState(k []byte) (bool, error) {
    _, ok := m.state[string(k)]
    return ok, nil
}

func (m *mockLedger) PrefixIterator(pref []byte) PrefixIterator {
    var items []KV
    for k, v := range m.state {
        if len(k) >= len(pref) && k[:len(pref)] == string(pref) {
            items = append(items, KV{k: []byte(k), v: v})
        }
    }
    return &sliceIter{items: items}
}

func (m *mockLedger) Transfer(from, to Address, amount uint64) error {
    m.transfers = append(m.transfers, from.String()+"->"+to.String())
    m.balances[to.String()] += amount
    if amount == 0 {
        return errors.New("zero")
    }
    return nil
}

func (m *mockLedger) BalanceOf(addr Address) uint64 {
    return m.balances[addr.String()]
}

// slice iterator implements PrefixIterator

type sliceIter struct {
    items []KV
    idx   int
}

func (s *sliceIter) Next() bool {
    if s.idx >= len(s.items) {
        return false
    }
    s.idx++
    return true
}

func (s *sliceIter) Key() []byte   { return s.items[s.idx-1].k }
func (s *sliceIter) Value() []byte { return s.items[s.idx-1].v }

// ------------------------------------------------------------
// mock electorate
// ------------------------------------------------------------

type mockElectorate struct{ holders map[Address]bool }

func (m mockElectorate) IsIDTokenHolder(a Address) bool { return m.holders[a] }

// ------------------------------------------------------------
// Tests
// ------------------------------------------------------------

func TestRegisterCharity(t *testing.T) {
    led := newMockLedger()
    cp := NewCharityPool(nil, led, mockElectorate{}, time.Now().UTC())
    addr := Address{0xAA}

    cases := []struct {
        name    string
        nameStr string
        cat     CharityCategory
        prefill bool
        expectErr bool
    }{
        {"OK", "GoodCharity", HungerRelief, false, false},
        {"Duplicate", "GoodCharity", HungerRelief, true, true},
        {"BadName", "", HungerRelief, false, true},
        {"BadCat", "Valid", CharityCategory(99), false, true},
    }

    for _, tc := range cases {
        if tc.prefill {
            _ = cp.Register(addr, tc.nameStr, tc.cat)
        }
        err := cp.Register(addr, tc.nameStr, tc.cat)
        if (err != nil) != tc.expectErr {
            t.Fatalf("%s: error expectation mismatch got %v", tc.name, err)
        }
        // cleanup state for next iteration
        led.state = make(map[string][]byte)
    }
}

func TestVoteLogic(t *testing.T) {
    led := newMockLedger()
    voter := Address{0x01}
    charity := Address{0x02}
    elect := mockElectorate{holders: map[Address]bool{voter: true}}

    genesis := time.Now().UTC().Add(-91 * 24 * time.Hour) // ensure within cycle
    cp := NewCharityPool(nil, led, elect, genesis)

    // first register charity for current cycle
    _ = cp.Register(charity, "Aid", HungerRelief)

    // fast-forward time into voting window (last 15d)
    timeNow = func() time.Time { return cp.cycleEnd(cp.currentCycle(time.Now().UTC())).Add(-votingWindow / 2) }
    defer func() { timeNow = time.Now }

    if err := cp.Vote(voter, charity); err != nil {
        t.Fatalf("unexpected vote error %v", err)
    }

    // duplicate vote
    if err := cp.Vote(voter, charity); err == nil {
        t.Fatalf("expected duplicate vote error")
    }

    // unverified voter
    badElect := mockElectorate{holders: map[Address]bool{}}
    cp.vote = badElect
    if err := cp.Vote(Address{0x03}, charity); err == nil {
        t.Fatalf("expected not verified voter error")
    }
}

// overrideable time function for tests
var timeNow = time.Now

func TestDailyPayout(t *testing.T) {
    led := newMockLedger()
    elect := mockElectorate{}
    genesis := time.Now().UTC().Add(-24 * time.Hour)
    cp := NewCharityPool(nil, led, elect, genesis)

    // setup winners list for cycle 0
    winners := []Address{{0x10}, {0x11}}
    raw, _ := json.Marshal(winners)
    led.SetState(winKey(0), raw)

    // fund charity pool balance
    led.balances[CharityPoolAccount.String()] = 1000

    cp.Tick(time.Now().UTC().Add(25 * time.Hour)) // 1 day later triggers payout

    if led.balances[InternalCharityAccount.String()] != 500 {
        t.Errorf("expected 500 to internal, got %d", led.balances[InternalCharityAccount.String()])
    }
    // each winner should get <=500/len(winners)
    per := uint64(500 / len(winners))
    for _, w := range winners {
        if led.balances[w.String()] != per {
            t.Errorf("winner balance wrong, want %d got %d", per, led.balances[w.String()])
        }
    }
}

func TestDeposit(t *testing.T) {
    led := newMockLedger()
    cp := NewCharityPool(nil, led, mockElectorate{}, time.Now())
    from := Address{0xDE}

    if err := cp.Deposit(from, 123); err != nil {
        t.Fatalf("deposit error: %v", err)
    }
    if len(led.transfers) != 1 {
        t.Fatalf("expected one transfer recorded")
    }
}
