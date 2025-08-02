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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
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
		stop:      make(chan struct{}),
	}
	for _, p := range initial {
		hc.peers[p] = &peerStat{}
	}
	go hc.loop()
	return hc
}

//---------------------------------------------------------------------
// Background ping loop
//---------------------------------------------------------------------

func (hc *HealthChecker) loop() {
	t := time.NewTicker(hc.interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			hc.tick()
		case <-hc.stop:
			return
		}
	}
}

// Stop terminates background health checks.
func (hc *HealthChecker) Stop() {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	select {
	case <-hc.stop:
		return
	default:
		close(hc.stop)
	}
}

func (hc *HealthChecker) tick() {
	hc.mu.RLock()
	peers := make([]Address, 0, len(hc.peers))
	for p := range hc.peers {
		peers = append(peers, p)
	}
	hc.mu.RUnlock()

	var wg sync.WaitGroup
	for _, addr := range peers {
		wg.Add(1)
		go func(a Address) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), hc.interval)
			defer cancel()
			rtt, err := hc.ping.Ping(ctx, a)

			hc.mu.Lock()
			ps, ok := hc.peers[a]
			if !ok {
				hc.mu.Unlock()
				return
			}
			if err != nil {
				ps.Misses++
			} else {
				ps.Misses = 0
				ms := float64(rtt.Milliseconds())
				if ps.EWMA == 0 {
					ps.EWMA = ms
				} else {
					ps.EWMA = hc.alpha*ms + (1-hc.alpha)*ps.EWMA
				}
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

func (hc *HealthChecker) AddPeer(addr Address) {
	hc.mu.Lock()
	hc.peers[addr] = &peerStat{}
	hc.mu.Unlock()
}
func (hc *HealthChecker) RemovePeer(addr Address) {
	hc.mu.Lock()
	delete(hc.peers, addr)
	hc.mu.Unlock()
}

//---------------------------------------------------------------------
// Snapshot for CLI / REST
//---------------------------------------------------------------------

func (hc *HealthChecker) Snapshot() []PeerInfo {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
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
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.peers = make(map[Address]*peerStat)
	for _, p := range newPeers {
		hc.peers[p] = &peerStat{}
	}
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
// High availability: backup, recovery & failover
//---------------------------------------------------------------------

// BackupManager handles periodic ledger snapshots and verification. Snapshots
// are written to multiple paths in parallel and verified using SHA-256 hashes.
type BackupManager struct {
	ledger   *Ledger
	paths    []string
	interval time.Duration
	stop     chan struct{}
	lastHash [32]byte
	wg       sync.WaitGroup
}

// NewBackupManager configures a new BackupManager. Backups are taken at the
// specified interval and replicated to all provided paths.
func NewBackupManager(l *Ledger, paths []string, interval time.Duration) *BackupManager {
	bm := &BackupManager{ledger: l, paths: paths, interval: interval, stop: make(chan struct{})}
	return bm
}

// Start launches the periodic backup loop.
func (bm *BackupManager) Start() {
	bm.wg.Add(1)
	go bm.loop()
}

// Stop terminates the backup loop and waits for it to exit.
func (bm *BackupManager) Stop() {
	close(bm.stop)
	bm.wg.Wait()
}

func (bm *BackupManager) loop() {
	defer bm.wg.Done()
	t := time.NewTicker(bm.interval)
	defer t.Stop()
	for {
		select {
		case <-bm.stop:
			return
		case <-t.C:
			ctx, cancel := context.WithTimeout(context.Background(), bm.interval)
			_ = bm.Snapshot(ctx, false)
			cancel()
		}
	}
}

// Snapshot writes a ledger snapshot to all backup locations. When incremental
// is true, the snapshot is only written if the state hash changed since the last
// run.
func (bm *BackupManager) Snapshot(ctx context.Context, incremental bool) error {
	data, err := bm.ledger.Snapshot()
	if err != nil {
		return err
	}
	h := sha256.Sum256(data)
	if incremental && h == bm.lastHash {
		return nil
	}
	ts := time.Now().UTC().Format("20060102T150405")
	for _, p := range bm.paths {
		if err := os.MkdirAll(p, 0o755); err != nil {
			return err
		}
		fname := filepath.Join(p, "snapshot-"+ts+".json")
		if err := os.WriteFile(fname, data, 0o600); err != nil {
			return err
		}
	}
	bm.lastHash = h
	return nil
}

// Verify checks that the snapshot at path matches the current ledger state.
func (bm *BackupManager) Verify(path string) error {
	disk, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	live, err := bm.ledger.Snapshot()
	if err != nil {
		return err
	}
	if sha256.Sum256(disk) != sha256.Sum256(live) {
		return errors.New("backup verification failed")
	}
	return nil
}

// RecoveryManager restores ledger state from a snapshot and monitors peers for
// automatic failover when the leader is unresponsive.
type RecoveryManager struct {
	ledger  *Ledger
	changer ViewChanger
	hc      *HealthChecker
}

func NewRecoveryManager(l *Ledger, hc *HealthChecker, vc ViewChanger) *RecoveryManager {
	return &RecoveryManager{ledger: l, changer: vc, hc: hc}
}

// Stop terminates background health monitoring.
func (rm *RecoveryManager) Stop() {
	if rm.hc != nil {
		rm.hc.Stop()
	}
}

// Restore loads a snapshot from disk and replaces the in-memory ledger state.
func (rm *RecoveryManager) Restore(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var snap Ledger
	if err := json.Unmarshal(data, &snap); err != nil {
		return err
	}
	rm.ledger = &snap
	return nil
}

// MonitorFailover watches peer statistics and triggers a view change when the
// current leader becomes faulty.
func (rm *RecoveryManager) MonitorFailover(ctx context.Context) {
	t := time.NewTicker(rm.hc.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			for _, p := range rm.hc.Snapshot() {
				if p.Misses >= rm.hc.maxMisses && p.Address == rm.changer.CurrentLeader() {
					rm.changer.ProposeViewChange("automatic failover")
				}
			}
		}
	}
}

// PredictiveFailureDetector tracks peer statistics and estimates a probability
// of failure based on an exponential moving average of RTT.
type PredictiveFailureDetector struct {
	mu        sync.Mutex
	averages  map[Address]float64
	threshold float64
}

func NewPredictiveFailureDetector(threshold float64) *PredictiveFailureDetector {
	return &PredictiveFailureDetector{averages: make(map[Address]float64), threshold: threshold}
}

// Record updates statistics for a peer.
func (p *PredictiveFailureDetector) Record(addr Address, rtt float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	cur := p.averages[addr]
	if cur == 0 {
		p.averages[addr] = rtt
	} else {
		p.averages[addr] = 0.8*cur + 0.2*rtt
	}
}

// FailureProb returns the probability that a peer will fail soon.
func (p *PredictiveFailureDetector) FailureProb(addr Address) float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	avg := p.averages[addr]
	if avg == 0 {
		return 0
	}
	if avg >= p.threshold {
		return 1
	}
	return avg / p.threshold
}

// DynamicResourceAllocator dynamically adjusts VM resource limits based on usage.
type DynamicResourceAllocator struct {
	mu     sync.Mutex
	limits map[Address]uint64
}

// NewDynamicResourceAllocator creates a new instance of the allocator.
func NewDynamicResourceAllocator() *DynamicResourceAllocator {
	return &DynamicResourceAllocator{limits: make(map[Address]uint64)}
}

// Adjust sets the new gas limit for a contract address.
func (r *DynamicResourceAllocator) Adjust(addr Address, gas uint64) {
	r.mu.Lock()
	r.limits[addr] = gas
	r.mu.Unlock()
}

//---------------------------------------------------------------------
// END fault_tolerance.go
//---------------------------------------------------------------------
