package core

import (
    "context"
    "errors"
    "sync"
    "testing"
    "time"
)

//------------------------------------------------------------
// mocks
//------------------------------------------------------------

type pingResp struct {
    dur time.Duration
    err error
}

type mockPinger struct {
    mu  sync.Mutex
    seq map[Address][]pingResp
}

func (m *mockPinger) Ping(ctx context.Context, a Address) (time.Duration, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    lst := m.seq[a]
    if len(lst) == 0 {
        return 0, errors.New("no response configured")
    }
    r := lst[0]
    m.seq[a] = lst[1:]
    return r.dur, r.err
}

type mockChanger struct {
    leader Address
    called bool
    reason string
}

func (m *mockChanger) CurrentLeader() Address { return m.leader }
func (m *mockChanger) ProposeViewChange(reason string) {
    m.called = true
    m.reason = reason
}

//------------------------------------------------------------
// helper to build HC without spinning goroutine
//------------------------------------------------------------

func newHC(interval time.Duration, p Pinger, c ViewChanger, peers []Address) *HealthChecker {
    hc := &HealthChecker{
        peers:     make(map[Address]*peerStat),
        interval:  interval,
        alpha:     0.2,
        maxRTT:    1500,
        maxMisses: 3,
        ping:      p,
        changer:   c,
    }
    for _, peer := range peers {
        hc.peers[peer] = &peerStat{}
    }
    return hc
}

//------------------------------------------------------------
// tests
//------------------------------------------------------------

func TestHealthChecker_TickScenarios(t *testing.T) {
    peer := Address{0xAA}

    cases := []struct {
        name        string
        pingSeq     []pingResp
        preMiss     int
        expectMiss  int
        expectEWMA  float64
        expectVCall bool
    }{
        {
            name:       "GoodPing",
            pingSeq:    []pingResp{{dur: 100 * time.Millisecond, err: nil}},
            preMiss:    0,
            expectMiss: 0,
            expectEWMA: 100,
            expectVCall: false,
        },
        {
            name:       "MissedPing",
            pingSeq:    []pingResp{{err: errors.New("fail")}},
            preMiss:    0,
            expectMiss: 1,
            expectEWMA: 0,
            expectVCall: false,
        },
        {
            name:       "FaultyLeaderTriggersViewChange",
            pingSeq:    []pingResp{{err: errors.New("fail")}},
            preMiss:    2, // will reach 3 => faulty
            expectMiss: 3,
            expectEWMA: 0,
            expectVCall: true,
        },
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            mp := &mockPinger{seq: map[Address][]pingResp{peer: tc.pingSeq}}
            mc := &mockChanger{leader: peer}
            hc := newHC(100*time.Millisecond, mp, mc, []Address{peer})

            st := hc.peers[peer]
            st.Misses = tc.preMiss

            hc.tick() // single evaluation

            if st.Misses != tc.expectMiss {
                t.Fatalf("miss count=%d want %d", st.Misses, tc.expectMiss)
            }
            if tc.expectEWMA > 0 && (st.EWMA < tc.expectEWMA-1 || st.EWMA > tc.expectEWMA+1) {
                t.Fatalf("EWMA=%f want %f", st.EWMA, tc.expectEWMA)
            }
            if mc.called != tc.expectVCall {
                t.Fatalf("view change called=%v want %v", mc.called, tc.expectVCall)
            }
        })
    }
}

func TestHealthChecker_Snapshot_Reconfigure(t *testing.T) {
    p1 := Address{0x01}
    p2 := Address{0x02}

    hc := newHC(time.Second, &mockPinger{seq: map[Address][]pingResp{}}, &mockChanger{}, []Address{p1, p2})

    snap := hc.Snapshot()
    if len(snap) != 2 {
        t.Fatalf("snapshot len=%d want 2", len(snap))
    }

    hc.Reconfigure([]Address{p2})
    if len(hc.peers) != 1 {
        t.Fatalf("peer set size %d want 1", len(hc.peers))
    }
    if _, ok := hc.peers[p2]; !ok {
        t.Fatal("expected p2 to remain after reconfigure")
    }
}
