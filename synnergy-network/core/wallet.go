package core

// Wallet implementation for the Synnergy Network blockchain.
//
// Features
// --------
//   * Ed25519 key‑pairs only (fast, deterministic and quantum‑resistant).
//   * Hierarchical Deterministic derivation (SLIP‑0010 / BIP‑32‑like).
//   * BIP‑39 mnemonic utilities (12‑/24‑word human recovery phrases).
//   * Address derivation (20‑byte SHA‑256/Ripemd‑160) matching the core address type.
//   * Transaction signing helper wired for core.Transaction & TxPool.
//   * Zero‑allocation logging and full error propagation – no placeholders.
//
// Import hygiene: wallet depends **only** on common + utility (crypto, log, bip‑libs).
// It does NOT import ledger, consensus, network or vm to stay at the lowest tier.

import (
	"crypto/ed25519"
	"crypto/hmac"
	crand "crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	bip39 "github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/ripemd160"
	"time"
)

//---------------------------------------------------------------------
// Constants and helpers
//---------------------------------------------------------------------

const (
	hardenedOffset uint32 = 0x80000000

	masterHMACKey = "ed25519 seed" // SLIP‑0010 master‑key string
)

func SetWalletLogger(l *log.Logger) { globalLogger = l }

var globalLogger = log.New()

//---------------------------------------------------------------------
// HDWallet structure
//---------------------------------------------------------------------

// HDWallet keeps master key material in‑memory only.
// *NEVER* persist the private fields directly – use encrypted keystores instead.
//
// Derivation model: SLIP‑0010 hardened children only, path m / account' / index'
// (change path omitted; wallets may overlay a change=1 hardened level if desired).
//
// This keeps derivation simple (ed25519 does not support unhardened children).

// Seed returns a copy of the wallet's master seed. Callers should securely wipe
// the returned slice after use.
func (w *HDWallet) Seed() []byte {
	out := make([]byte, len(w.seed))
	copy(out, w.seed)
	return out
}

//---------------------------------------------------------------------
// Wallet creation utilities
//---------------------------------------------------------------------

// NewRandomWallet generates entropyBits (128/256) of RNG entropy, returns wallet + mnemonic.
// The caller MUST wipe the mnemonic or store it securely.
func NewRandomWallet(entropyBits int) (*HDWallet, string, error) {
	if entropyBits != 128 && entropyBits != 256 {
		return nil, "", fmt.Errorf("unsupported entropy size %d", entropyBits)
	}

	entropy, err := bip39.NewEntropy(entropyBits)
	if err != nil {
		return nil, "", fmt.Errorf("entropy: %w", err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, "", fmt.Errorf("mnemonic: %w", err)
	}
	seed := bip39.NewSeed(mnemonic, "")
	w, err := NewHDWalletFromSeed(seed, globalLogger)
	if err != nil {
		return nil, "", err
	}
	return w, mnemonic, nil
}

// WalletFromMnemonic imports an existing BIP‑39 phrase.
func WalletFromMnemonic(mnemonic, passphrase string) (*HDWallet, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("invalid mnemonic checksum")
	}
	seed := bip39.NewSeed(mnemonic, passphrase)
	return NewHDWalletFromSeed(seed, globalLogger)
}

func NewHDWalletFromSeed(seed []byte, lg *log.Logger) (*HDWallet, error) {
	if len(seed) < 16 {
		return nil, errors.New("seed too short")
	}

	I := hmacSHA512([]byte(masterHMACKey), seed)

	w := &HDWallet{
		seed:        seed,
		masterKey:   I[:32],
		masterChain: I[32:],
		logger:      lg,
	}

	lg.Infof("wallet: master key initialised (%d bytes seed)", len(seed))
	return w, nil
}

//---------------------------------------------------------------------
// Derivation path helpers
//---------------------------------------------------------------------

// derivePrivate returns the key material & new chain‑code for a (hardened) index.
// Only hardened derivation is supported for ed25519 – index MUST already carry the hardened offset.
func derivePrivate(parentKey, parentChain []byte, index uint32) (key, ccode []byte, err error) {
	if index < hardenedOffset {
		return nil, nil, errors.New("non‑hardened derivation not supported for ed25519")
	}
	// Data = 0x00 || parentKey || index(be)
	data := make([]byte, 1+32+4)
	copy(data[1:], parentKey)
	binary.BigEndian.PutUint32(data[33:], index)

	I := hmacSHA512(parentChain, data)
	key = I[:32]
	ccode = I[32:]
	return key, ccode, nil
}

// HMAC‑SHA512 helper (constant‑time)
func hmacSHA512(key, data []byte) []byte {
	h := hmac.New(sha512.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// PrivateKey returns the (ed25519) private key for derivation path m / account' / index'.
// account, index are hardened internally.
func (w *HDWallet) PrivateKey(account, index uint32) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	account |= hardenedOffset
	index |= hardenedOffset

	// First level: account'
	k1, c1, err := derivePrivate(w.masterKey, w.masterChain, account)
	if err != nil {
		return nil, nil, err
	}
	// Second level: index'
	k2, _, err := derivePrivate(k1, c1, index)
	if err != nil {
		return nil, nil, err
	}
	priv := ed25519.NewKeyFromSeed(k2)       // 64‑byte private key (seed+pub)
	pub := priv.Public().(ed25519.PublicKey) // 32‑byte
	return priv, pub, nil
}

//---------------------------------------------------------------------
// Address helpers
//---------------------------------------------------------------------

// Address converts a 32‑byte ed25519 public key into 20‑byte account address.
// The scheme: SHA‑256(pub) → RIPEMD‑160 → Address.
func pubKeyToAddress(pub ed25519.PublicKey) Address {
	sha := sha256.Sum256(pub)
	ripemd := ripemd160.New()
	ripemd.Write(sha[:])
	var out Address
	copy(out[:], ripemd.Sum(nil))
	return out
}

// NewAddress derives account+index and returns its Address.
func (w *HDWallet) NewAddress(account, index uint32) (Address, error) {
	_, pub, err := w.PrivateKey(account, index)
	if err != nil {
		return Address{}, err
	}
	return pubKeyToAddress(pub), nil
}

//---------------------------------------------------------------------
// Transaction signing
//---------------------------------------------------------------------

// SignTx derives (account, index) key, signs tx, sets tx.Sig and From.
// Signature layout: [64‑byte sig || 32‑byte pubkey] to allow stateless verification.
//
// The SignTx helper fully recalculates tx.Hash (double‑SHA256 in core.transactions.go).
func (w *HDWallet) SignTx(tx *Transaction, account, index uint32, gasPrice uint64) error {
	if tx == nil {
		return errors.New("nil transaction")
	}
	priv, pub, err := w.PrivateKey(account, index)
	if err != nil {
		return err
	}
	addr := pubKeyToAddress(pub)

	// Attach sender & gas price beforehand for hashing.
	tx.From = addr
	if gasPrice > 0 {
		tx.GasPrice = gasPrice
	}
	tx.Timestamp = time.Now().UnixMilli()

	hash := tx.HashTx()

	sig := ed25519.Sign(priv, hash[:])

	signed := make([]byte, 96)
	copy(signed[:64], sig)
	copy(signed[64:], pub)
	tx.Sig = signed

	w.logger.Printf("signed tx %s by %s (account %d idx %d)", hash.Short(), addr.Short(), account, index)
	return nil
}

//---------------------------------------------------------------------
// Utility helpers
//---------------------------------------------------------------------

// RandomMnemonicEntropy produces cryptographically‑secure random entropy of given bits.
func RandomMnemonicEntropy(bits int) ([]byte, error) {
	if bits%32 != 0 {
		return nil, errors.New("entropy bits must be multiple of 32")
	}
	b := make([]byte, bits/8)
	if _, err := crand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// Wipe zeroes a byte slice in‑place (best‑effort – GC might still copy).
func Wipe(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// Hex returns the full hexadecimal representation of the address.
func (a Address) Hex() string {
	return "0x" + hex.EncodeToString(a[:])
}

// Short returns a shortened version (first 4 + last 4 hex chars).
func (a Address) Short() string {
	full := hex.EncodeToString(a[:])
	if len(full) <= 8 {
		return full
	}
	return fmt.Sprintf("%s..%s", full[:4], full[len(full)-4:])
}
