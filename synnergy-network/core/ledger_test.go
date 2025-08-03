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
			config, _ := tmpLedgerConfig(t, tc.genesis)
			ledger, err := NewLedger(config)
			if err != nil {
				t.Fatalf("init err: %v", err)
			}
			if len(ledger.Blocks) != tc.wantBlocks {
				t.Fatalf("blocks=%d want %d", len(ledger.Blocks), tc.wantBlocks)
			}
		})
	}
}

//-------------------------------------------------------------
// Test AddBlock height validation
//-------------------------------------------------------------

func TestAddBlockHeightMismatch(t *testing.T) {
	genesis := &Block{Header: BlockHeader{Height: 0}}
	config, _ := tmpLedgerConfig(t, genesis)
	ledger, _ := NewLedger(config)

	// create block with incorrect height (should be 1)
	badBlock := &Block{Header: BlockHeader{Height: 2}}
	if err := ledger.AddBlock(badBlock); err == nil {
		t.Fatalf("expected height mismatch error")
	}
}

//-------------------------------------------------------------
// Test MintToken and BalanceOf
//-------------------------------------------------------------

func TestMintTokenBalance(t *testing.T) {
	cfg, _ := tmpLedgerConfig(t, nil)
	ledger, _ := NewLedger(cfg)
	addr := Address{0xAA}

	if err := ledger.MintToken(addr, "SYNN", 0); err == nil {
		t.Fatalf("expected zero amount error")
	}
	if err := ledger.MintToken(addr, "SYNN", 500); err != nil {
		t.Fatalf("mint err %v", err)
	}
	bal := ledger.BalanceOf(addr)
	if bal != 500 {
		t.Fatalf("balance %d want 500", bal)
	}
}

//-------------------------------------------------------------
// Test Snapshot round‑trip
//-------------------------------------------------------------

func TestSnapshotRoundTrip(t *testing.T) {
	config, _ := tmpLedgerConfig(t, nil)
	ledger, _ := NewLedger(config)
	ledger.State["foo"] = []byte("bar")
	data, err := ledger.Snapshot()
	if err != nil {
		t.Fatalf("snapshot err %v", err)
	}

	var outLedger Ledger
	if err := json.Unmarshal(data, &outLedger); err != nil {
		t.Fatalf("unmarshal snapshot %v", err)
	}
	if val := outLedger.State["foo"]; string(val) != "bar" {
		t.Fatalf("snapshot state mismatch")
	}
}

//-------------------------------------------------------------
// Test AppendSubBlock continuity rule
//-------------------------------------------------------------

func TestAppendSubBlockHeightCheck(t *testing.T) {
	// bootstrap ledger with one block that has no subheaders yet
	block := &Block{Header: BlockHeader{Height: 0}, Body: BlockBody{SubHeaders: []SubBlockHeader{}}}
	config, _ := tmpLedgerConfig(t, block)
	ledger, _ := NewLedger(config)

	// first sub-block height 0 ok
	sbZero := &SubBlock{Header: SubBlockHeader{Height: 0}}
	sbZero.Body.Transactions = [][]byte{[]byte("tx")}
	if err := ledger.AppendSubBlock(sbZero); err != nil {
		t.Fatalf("append subblock 0 err %v", err)
	}

	// next sub-block with same height should fail
	sbDuplicate := &SubBlock{Header: SubBlockHeader{Height: 0}}
	if err := ledger.AppendSubBlock(sbDuplicate); err == nil {
		t.Fatalf("expected height mismatch error")
	}
}

//-------------------------------------------------------------
// Test pruning archives old blocks
//-------------------------------------------------------------

func TestPruneArchivesBlocks(t *testing.T) {
	genesis := &Block{Header: BlockHeader{Height: 0}}
	config, cleanup := tmpLedgerConfig(t, genesis)
	defer cleanup()
	config.PruneInterval = 2
	ledger, err := NewLedger(config)
	if err != nil {
		t.Fatalf("ledger init: %v", err)
	}

	// add blocks 1,2,3 - block 0 should be pruned
	for i := 1; i <= 3; i++ {
		block := &Block{Header: BlockHeader{Height: uint64(i)}}
		if err := ledger.AddBlock(block); err != nil {
			t.Fatalf("add block %d: %v", i, err)
		}
	}

	if got := len(ledger.Blocks); got != 2 {
		t.Fatalf("expected 2 blocks after prune, got %d", got)
	}

	// ensure archive file has data
	info, err := os.Stat(config.ArchivePath)
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
	config, cleanup := tmpLedgerConfig(t, nil)
	defer cleanup()
	ledgerA, _ := NewLedger(config)
	ledgerA.State["a"] = []byte("1")
	ledgerA.State["b"] = []byte("2")

	config2, cleanup2 := tmpLedgerConfig(t, nil)
	defer cleanup2()
	ledgerB, _ := NewLedger(config2)
	ledgerB.State["b"] = []byte("2")
	ledgerB.State["a"] = []byte("1")

	if ledgerA.StateRoot() != ledgerB.StateRoot() {
		t.Fatalf("state roots mismatch")
	}
}
