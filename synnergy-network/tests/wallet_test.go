package core_test

import (
	"bytes"
	"encoding/hex"
	core "synnergy-network/core"
	"testing"
)

// ------------------------------------------------------------
// Test NewRandomWallet & WalletFromMnemonic
// ------------------------------------------------------------

func TestNewRandomWalletAndImport(t *testing.T) {
	cases := []struct {
		name    string
		entropy int
		wantErr bool
	}{
		{"Entropy128", 128, false},
		{"Entropy256", 256, false},
		{"BadEntropy", 192, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w, m, err := NewRandomWallet(tc.entropy)
			if (err != nil) != tc.wantErr {
				t.Fatalf("unexpected error status: %v", err)
			}
			if tc.wantErr {
				return
			}
			if w == nil || len(m) == 0 {
				t.Fatalf("wallet or mnemonic not returned")
			}

			// round‑trip import
			w2, err := WalletFromMnemonic(m, "")
			if err != nil {
				t.Fatalf("import failed: %v", err)
			}
			// derive first address from both wallets and compare
			addr1, _ := w.NewAddress(0, 0)
			addr2, _ := w2.NewAddress(0, 0)
			if addr1 != addr2 {
				t.Errorf("addresses mismatch after import")
			}
		})
	}
}

// ------------------------------------------------------------
// Test PrivateKey & NewAddress deterministic output
// ------------------------------------------------------------

func TestPrivateKeyDeterministic(t *testing.T) {
	w, _, _ := NewRandomWallet(128)
	priv1, pub1, err := w.PrivateKey(1, 2)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	priv2, pub2, _ := w.PrivateKey(1, 2)

	if !bytes.Equal(priv1, priv2) || !bytes.Equal(pub1, pub2) {
		t.Errorf("deterministic derivation failed")
	}

	if len(priv1) != 64 || len(pub1) != 32 {
		t.Errorf("unexpected key sizes priv=%d pub=%d", len(priv1), len(pub1))
	}

	// Address derivation consistency
	addr, _ := w.NewAddress(1, 2)
	expected := hex.EncodeToString(pub1)      // just ensure address derived from pub key differs in size
	if len(addr) == 0 || len(expected) == 0 { // dummy sanity
		t.Errorf("address conversion failed")
	}
}

// ------------------------------------------------------------
// Test RandomMnemonicEntropy helper
// ------------------------------------------------------------

func TestRandomMnemonicEntropy(t *testing.T) {
	if _, err := RandomMnemonicEntropy(130); err == nil {
		t.Fatalf("expected bits multiple of 32 error")
	}
	out, err := RandomMnemonicEntropy(256)
	if err != nil || len(out) != 32 {
		t.Fatalf("bad entropy: %v len %d", err, len(out))
	}
}

// ------------------------------------------------------------
// Test SignTx
// ------------------------------------------------------------

func TestSignTx(t *testing.T) {
	w, _, _ := NewRandomWallet(128)
	tx := &Transaction{}

	// nil tx error
	if err := w.SignTx(nil, 0, 0, 0); err == nil {
		t.Fatalf("expected error on nil tx")
	}

	err := w.SignTx(tx, 0, 1, 1000)
	if err != nil {
		t.Fatalf("sign tx err: %v", err)
	}

	if len(tx.Sig) != 96 {
		t.Errorf("signature length want 96 got %d", len(tx.Sig))
	}
	if (tx.From == Address{}) {
		t.Errorf("tx.From not set")
	}
}

// ------------------------------------------------------------
// Test Wipe utility
// ------------------------------------------------------------

func TestWipe(t *testing.T) {
	b := []byte{1, 2, 3}
	Wipe(b)
	for _, v := range b {
		if v != 0 {
			t.Errorf("wipe failed value %d", v)
		}
	}
}
