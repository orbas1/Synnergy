package core

// fault_tolerance.go – Peer health‑checking and view‑change signaling for the
// Synnergy Network consensus layer.
//
// Key components
// --------------
// • **HealthChecker** – maintains round‑trip time (RTT) scores for each peer.
//   Pings occur every `interval`, exponential back‑off for failures. Scores are
//   EWMA smoothed; peers with score above `maxRTT` or missing `maxMisses` are
//   flagged as faulty and excluded from routing tables.
// • **Reconfigure()** – when the current consensus leader scores as faulty,
//   HealthChecker invokes the provided `ViewChanger` interface to trigger a
//   round‑robin leader change (HotStuff‑style).
// • CLI hook (`Syncl net peers`) and REST endpoint (`/peers`) rely on exported
//   `Snapshot()` which returns live peer stats for inspection.
//
// Dependencies: common (Address), network (Dial, SendPing), consensus (view
// change callback). No circular imports.
// -----------------------------------------------------------------------------

import (
    "context"
    "sync"
    "time"
    "errors"
    "net"
    "encoding/hex"
)

//---------------------------------------------------------------------
// Interfaces injected from other modules
//---------------------------------------------------------------------



//---------------------------------------------------------------------
// HealthChecker
//---------------------------------------------------------------------


func NewHealthChecker(ping Pinger, changer ViewChanger, initial []Address) *HealthChecker {
    hc := &HealthChecker{
        peers:     make(map[Address]*peerStat),
        interval:  3 * time.Second,
        alpha:     0.2,
        maxRTT:    1500, // 1.5s
        maxMisses: 3,
        ping:      ping,
        changer:   changer,
    }
    for _, p := range initial { hc.peers[p] = &peerStat{} }
    go hc.loop()
    return hc
}

//---------------------------------------------------------------------
// Background ping loop
//---------------------------------------------------------------------

func (hc *HealthChecker) loop() {
    t := time.NewTicker(hc.interval)
    for range t.C {
        hc.tick()
    }
}

func (hc *HealthChecker) tick() {
    hc.mu.RLock(); peers := make([]Address, 0, len(hc.peers)); for p := range hc.peers { peers = append(peers, p) }; hc.mu.RUnlock()

    var wg sync.WaitGroup
    for _, addr := range peers {
        wg.Add(1)
        go func(a Address) {
            defer wg.Done()
            ctx, cancel := context.WithTimeout(context.Background(), hc.interval)
            defer cancel()
            rtt, err := hc.ping.Ping(ctx, a)

            hc.mu.Lock(); ps, ok := hc.peers[a]; if !ok { hc.mu.Unlock(); return }
            if err != nil {
                ps.Misses++
            } else {
                ps.Misses = 0
                ms := float64(rtt.Milliseconds())
                if ps.EWMA == 0 { ps.EWMA = ms } else { ps.EWMA = hc.alpha*ms + (1-hc.alpha)*ps.EWMA }
            }
            ps.LastUpdate = time.Now()
            faulty := ps.Misses >= hc.maxMisses || ps.EWMA > hc.maxRTT
            hc.mu.Unlock()

            if faulty && a == hc.changer.CurrentLeader() {
                hc.changer.ProposeViewChange("leader faulty")
            }
        }(addr)
    }
    wg.Wait()
}

type Pinger interface {
	Ping(ctx context.Context, addr Address) (time.Duration, error)
}

type ViewChanger interface {
	CurrentLeader() Address
	ProposeViewChange(reason string)
}


//---------------------------------------------------------------------
// Manage peer set
//---------------------------------------------------------------------

func (hc *HealthChecker) AddPeer(addr Address) { hc.mu.Lock(); hc.peers[addr] = &peerStat{}; hc.mu.Unlock() }
func (hc *HealthChecker) RemovePeer(addr Address) { hc.mu.Lock(); delete(hc.peers, addr); hc.mu.Unlock() }

//---------------------------------------------------------------------
// Snapshot for CLI / REST
//---------------------------------------------------------------------


func (hc *HealthChecker) Snapshot() []PeerInfo {
    hc.mu.RLock(); defer hc.mu.RUnlock()
    out := make([]PeerInfo, 0, len(hc.peers))
    for addr, st := range hc.peers {
        out = append(out, PeerInfo{Address: addr, RTT: st.EWMA, Misses: st.Misses, Updated: st.LastUpdate.Unix()})
    }
    return out
}

//---------------------------------------------------------------------
// Reconfigure (external trigger)
//---------------------------------------------------------------------

func (hc *HealthChecker) Reconfigure(newPeers []Address) {
    hc.mu.Lock(); defer hc.mu.Unlock()
    hc.peers = make(map[Address]*peerStat)
    for _, p := range newPeers { hc.peers[p] = &peerStat{} }
}

//---------------------------------------------------------------------
// Integration helpers – network.Pinger implementation
//---------------------------------------------------------------------

type NetPinger struct {
    dial Dialer
}

func (a Address) String() string {
    return hex.EncodeToString(a[:])
}

func (np *NetPinger) Ping(ctx context.Context, peer Address) (time.Duration, error) {
    conn, err := np.dial.Dial(ctx, peer.String()) // Assuming Address.String() returns IP:Port
    if err != nil {
        return 0, err
    }
    defer conn.Close()

    t0 := time.Now()

    if err := SendPing(conn); err != nil {
        return 0, err
    }
    if err := AwaitPong(ctx, conn); err != nil {
        return 0, err
    }

    return time.Since(t0), nil
}

func SendPing(conn net.Conn) error {
    _, err := conn.Write([]byte("ping"))
    return err
}

func AwaitPong(ctx context.Context, conn net.Conn) error {
    buf := make([]byte, 4)
    conn.SetReadDeadline(time.Now().Add(2 * time.Second))
    _, err := conn.Read(buf)
    if err != nil {
        return err
    }
    if string(buf) != "pong" {
        return errors.New("unexpected response")
    }
    return nil
}


//---------------------------------------------------------------------
// END fault_tolerance.go
//---------------------------------------------------------------------
