package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// ------------------------------------------------------------
// Helper to create temporary ledger configuration for tests
// ------------------------------------------------------------

func tmpLedgerConfig(t *testing.T, genesis *Block) (LedgerConfig, func()) {
	dir := t.TempDir()
	wal := filepath.Join(dir, "wal.log")
	snap := filepath.Join(dir, "snap.json")
	arch := filepath.Join(dir, "archive.gz")
	cfg := LedgerConfig{
		WALPath:          wal,
		SnapshotPath:     snap,
		SnapshotInterval: 1000, // large to avoid snapshot during tests
		GenesisBlock:     genesis,
		ArchivePath:      arch,
	}
	cleanup := func() { os.RemoveAll(dir) }
	return cfg, cleanup
}

//-------------------------------------------------------------
// Test NewLedger with and without genesis
//-------------------------------------------------------------

func TestNewLedgerInit(t *testing.T) {
	tests := []struct {
		name       string
		genesis    *Block
		wantBlocks int
	}{
		{"Empty", nil, 0},
		{"WithGenesis", &Block{Header: BlockHeader{Height: 0}}, 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg, _ := tmpLedgerConfig(t, tc.genesis)
			led, err := NewLedger(cfg)
			if err != nil {
				t.Fatalf("init err: %v", err)
			}
			if len(led.Blocks) != tc.wantBlocks {
				t.Fatalf("blocks=%d want %d", len(led.Blocks), tc.wantBlocks)
			}
		})
	}
}

//-------------------------------------------------------------
// Test AddBlock height validation
//-------------------------------------------------------------

func TestAddBlockHeightMismatch(t *testing.T) {
	genesis := &Block{Header: BlockHeader{Height: 0}}
	cfg, _ := tmpLedgerConfig(t, genesis)
	led, _ := NewLedger(cfg)

	// create block with incorrect height (should be 1)
	bad := &Block{Header: BlockHeader{Height: 2}}
	if err := led.AddBlock(bad); err == nil {
		t.Fatalf("expected height mismatch error")
	}
}

//-------------------------------------------------------------
// Test MintToken and BalanceOf
//-------------------------------------------------------------

func TestMintTokenBalance(t *testing.T) {
	cfg, _ := tmpLedgerConfig(t, nil)
	led, _ := NewLedger(cfg)
	addr := Address{0xAA}

	if err := led.MintToken(addr, "SYNN", 0); err == nil {
		t.Fatalf("expected zero amount error")
	}
	if err := led.MintToken(addr, "SYNN", 500); err != nil {
		t.Fatalf("mint err %v", err)
	}
	bal := led.BalanceOf(addr)
	if bal != 500 {
		t.Fatalf("balance %d want 500", bal)
	}
}

//-------------------------------------------------------------
// Test Snapshot round‑trip
//-------------------------------------------------------------

func TestSnapshotRoundTrip(t *testing.T) {
	cfg, _ := tmpLedgerConfig(t, nil)
	led, _ := NewLedger(cfg)
	led.State["foo"] = []byte("bar")
	data, err := led.Snapshot()
	if err != nil {
		t.Fatalf("snapshot err %v", err)
	}

	var out Ledger
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal snapshot %v", err)
	}
	if val := out.State["foo"]; string(val) != "bar" {
		t.Fatalf("snapshot state mismatch")
	}
}

//-------------------------------------------------------------
// Test AppendSubBlock continuity rule
//-------------------------------------------------------------

func TestAppendSubBlockHeightCheck(t *testing.T) {
	// bootstrap ledger with one block that has no subheaders yet
	blk := &Block{Header: BlockHeader{Height: 0}, Body: BlockBody{SubHeaders: []SubBlockHeader{}}}
	cfg, _ := tmpLedgerConfig(t, blk)
	led, _ := NewLedger(cfg)

	// first sub-block height 0 ok
	sb0 := &SubBlock{Header: SubBlockHeader{Height: 0}}
	sb0.Body.Transactions = [][]byte{[]byte("tx")}
	if err := led.AppendSubBlock(sb0); err != nil {
		t.Fatalf("append subblock 0 err %v", err)
	}

	// next sub-block with same height should fail
	sbBad := &SubBlock{Header: SubBlockHeader{Height: 0}}
	if err := led.AppendSubBlock(sbBad); err == nil {
		t.Fatalf("expected height mismatch error")
	}
}

//-------------------------------------------------------------
// Test pruning archives old blocks
//-------------------------------------------------------------

func TestPruneArchivesBlocks(t *testing.T) {
	genesis := &Block{Header: BlockHeader{Height: 0}}
	cfg, cleanup := tmpLedgerConfig(t, genesis)
	defer cleanup()
	cfg.PruneInterval = 2
	led, err := NewLedger(cfg)
	if err != nil {
		t.Fatalf("ledger init: %v", err)
	}

	// add blocks 1,2,3 - block 0 should be pruned
	for i := 1; i <= 3; i++ {
		blk := &Block{Header: BlockHeader{Height: uint64(i)}}
		if err := led.AddBlock(blk); err != nil {
			t.Fatalf("add block %d: %v", i, err)
		}
	}

	if got := len(led.Blocks); got != 2 {
		t.Fatalf("expected 2 blocks after prune, got %d", got)
	}

	// ensure archive file has data
	info, err := os.Stat(cfg.ArchivePath)
	if err != nil {
		t.Fatalf("archive stat: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("archive file empty")
	}
}

//-------------------------------------------------------------
// Test StateRoot determinism
//-------------------------------------------------------------

func TestStateRootDeterministic(t *testing.T) {
	cfg, cleanup := tmpLedgerConfig(t, nil)
	defer cleanup()
	ledA, _ := NewLedger(cfg)
	ledA.State["a"] = []byte("1")
	ledA.State["b"] = []byte("2")

	cfg2, cleanup2 := tmpLedgerConfig(t, nil)
	defer cleanup2()
	ledB, _ := NewLedger(cfg2)
	ledB.State["b"] = []byte("2")
	ledB.State["a"] = []byte("1")

	if ledA.StateRoot() != ledB.StateRoot() {
		t.Fatalf("state roots mismatch")
	}
}
