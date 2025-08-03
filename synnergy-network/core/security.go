// SPDX-License-Identifier: Apache-2.0
// Package core – shared security primitives for the Synnergy Network stack.
//
// Exposes:
//   - Sign / Verify      – Ed25519 (wallets) + BLS12-381 (validators).
//   - BLS aggregation    – multi-sig / threshold helpers.
//   - XChaCha20-Poly1305 – authenticated encryption.
//   - ComputeMerkleRoot – Bitcoin-style double-SHA256 Merkle tree.
//   - TLS loader         – hardened TLS 1.3 config for node-to-node gRPC.
//
// All crypto comes from Go 1.22 std-lib or herumi BLS (battle-tested).
package core

import (
	"bufio"
	"bytes"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	mode3 "github.com/cloudflare/circl/sign/dilithium/mode3"
	bls "github.com/herumi/bls-eth-go-binary/bls"

	"golang.org/x/crypto/chacha20poly1305"
)

//---------------------------------------------------------------------
// Package-level init – BLS curve setup
//---------------------------------------------------------------------

func init() {
	if err := bls.Init(bls.BLS12_381); err != nil {
		panic(fmt.Errorf("bls init: %w", err))
	}
}

//---------------------------------------------------------------------
// Logger
//---------------------------------------------------------------------

var secLogger = log.New(io.Discard, "[security] ", log.LstdFlags)

func SetSecurityLogger(l *log.Logger) { secLogger = l }

//---------------------------------------------------------------------
// Sign / Verify – Ed25519 (default) & BLS12-381 (validators)
//---------------------------------------------------------------------

type KeyAlgo uint8

const (
	AlgoEd25519 KeyAlgo = iota
	AlgoBLS
)

// Sign signs msg with priv.
// - For Ed25519: priv must be ed25519.PrivateKey.
// - For BLS:     priv must be *bls.SecretKey.
func Sign(algo KeyAlgo, priv interface{}, msg []byte) ([]byte, error) {
	switch algo {
	case AlgoEd25519:
		pk, ok := priv.(ed25519.PrivateKey)
		if !ok {
			return nil, errors.New("invalid ed25519 private key type")
		}
		return ed25519.Sign(pk, msg), nil

	case AlgoBLS:
		sk, ok := priv.(*bls.SecretKey)
		if !ok {
			return nil, errors.New("invalid BLS secret key type")
		}
		sig := sk.SignByte(msg) // *bls.Sign
		return sig.Serialize(), nil

	default:
		return nil, errors.New("unknown algo")
	}
}

// Verify checks sig for msg with pub.
// pub may be ed25519.PublicKey, *bls.PublicKey, or compressed []byte (BLS).
func Verify(algo KeyAlgo, pub interface{}, msg, sig []byte) (bool, error) {
	switch algo {
	case AlgoEd25519:
		pk, ok := pub.(ed25519.PublicKey)
		if !ok {
			return false, errors.New("invalid ed25519 pubkey type")
		}
		return ed25519.Verify(pk, msg, sig), nil

	case AlgoBLS:
		var pk bls.PublicKey
		switch v := pub.(type) {
		case *bls.PublicKey:
			pk = *v
		case []byte:
			if err := pk.Deserialize(v); err != nil {
				return false, err
			}
		default:
			return false, errors.New("invalid BLS pubkey type")
		}

		var s bls.Sign
		if err := s.Deserialize(sig); err != nil {
			return false, err
		}
		return s.VerifyByte(&pk, msg), nil

	default:
		return false, errors.New("unknown algo")
	}
}

//---------------------------------------------------------------------
// BLS aggregation helpers
//---------------------------------------------------------------------

// AggregateBLSSigs merges multiple **compressed** BLS signatures.
func AggregateBLSSigs(sigs [][]byte) ([]byte, error) {
	if len(sigs) == 0 {
		return nil, errors.New("no sigs to aggregate")
	}
	var agg bls.Sign
	for i, raw := range sigs {
		var s bls.Sign
		if err := s.Deserialize(raw); err != nil {
			return nil, fmt.Errorf("sig %d: %w", i, err)
		}
		if i == 0 {
			agg = s
		} else {
			agg.Add(&s)
		}
	}
	return agg.Serialize(), nil
}

// VerifyAggregated verifies an aggregated sig for identical msg.
func VerifyAggregated(aggSig, pubAgg, msg []byte) (bool, error) {
	var pk bls.PublicKey
	if err := pk.Deserialize(pubAgg); err != nil {
		return false, err
	}
	var sig bls.Sign
	if err := sig.Deserialize(aggSig); err != nil {
		return false, err
	}
	return sig.VerifyByte(&pk, msg), nil
}

//---------------------------------------------------------------------
// Simple threshold reconstruction (Shamir over GF(256)) – Ed25519 seeds
//---------------------------------------------------------------------

type Share struct {
	Index byte   // 1-based index
	Data  []byte // 32-byte seed share
}

func CombineShares(shares []Share, threshold int) ([]byte, error) {
	if len(shares) < threshold {
		return nil, errors.New("not enough shares")
	}
	secret := make([]byte, 32)
	for i := 0; i < threshold; i++ {
		li := lagrangeCoeff(i, shares[:threshold])
		for b := 0; b < 32; b++ {
			secret[b] ^= gfMul(li, shares[i].Data[b])
		}
	}
	return secret, nil
}

func lagrangeCoeff(i int, ss []Share) byte {
	xi := ss[i].Index
	num, den := byte(1), byte(1)
	for j, s := range ss {
		if j == i {
			continue
		}
		xj := s.Index
		num = gfMul(num, xj)
		den = gfMul(den, xj^xi)
	}
	return gfDiv(num, den)
}

// GF(256) helpers (irreducible poly 0x11B)
func gfMul(a, b byte) byte {
	var p byte
	for b > 0 {
		if b&1 == 1 {
			p ^= a
		}
		hi := a & 0x80
		a <<= 1
		if hi != 0 {
			a ^= 0x1B
		}
		b >>= 1
	}
	return p
}

// Multiplicative inverse in GF(2⁸) using the extended Euclidean algorithm.
func gfInv(a byte) byte {
	if a == 0 {
		panic("inverse of zero")
	}
	var t0, t1 uint16 = 0, 1
	r0, r1 := uint16(0x11B), uint16(a) // 0x11B fits in 16 bits

	for r1 != 0 {
		q := polyDiv16(r0, r1)                           // ← 16-bit helpers
		r0, r1 = r1, r0^uint16(gfMul(byte(q), byte(r1))) // promote to uint16
		t0, t1 = t1, t0^uint16(gfMul(byte(q), byte(t1)))
	}
	return byte(t0)
}

func polyDiv16(a, b uint16) uint16 { // very small & slow, but fine for 8-bit field
	for shift := 15; shift >= 0; shift-- {
		if (b<<shift)&0xFF00 == a&0xFF00 { // match high byte
			return 1 << shift
		}
	}
	return 0
}

func gfDiv(a, b byte) byte { return gfMul(a, gfInv(b)) }
func polyDiv(a, b byte) byte {
	for i := 0; i < 8; i++ {
		if b<<i == a {
			return 1 << i
		}
	}
	return 0
}

//---------------------------------------------------------------------
// Encryption – XChaCha20-Poly1305
//---------------------------------------------------------------------

// Encrypt returns nonce || ciphertext || tag using XChaCha20-Poly1305.
func Encrypt(key, plaintext, aad []byte) ([]byte, error) {
	if len(key) != chacha20poly1305.KeySize { // ← use KeySize
		return nil, errors.New("key must be 32 bytes")
	}
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, chacha20poly1305.NonceSizeX) // 24-byte nonce
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ct := aead.Seal(nil, nonce, plaintext, aad)
	return append(nonce, ct...), nil
}

// Decrypt verifies and opens a blob produced by Encrypt.
func Decrypt(key, blob, aad []byte) ([]byte, error) {
	if len(key) != chacha20poly1305.KeySize { // ← use KeySize
		return nil, errors.New("key must be 32 bytes")
	}
	minLen := chacha20poly1305.NonceSizeX + chacha20poly1305.Overhead
	if len(blob) < minLen {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := blob[:chacha20poly1305.NonceSizeX], blob[chacha20poly1305.NonceSizeX:]
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}
	return aead.Open(nil, nonce, ciphertext, aad)
}

//---------------------------------------------------------------------
// Merkle root (double-SHA256, canonical ordering)
//---------------------------------------------------------------------

func ComputeMerkleRoot(leaves [][]byte) ([]byte, error) {
	if len(leaves) == 0 {
		return nil, errors.New("no leaves")
	}
	sort.SliceStable(leaves, func(i, j int) bool { return bytes.Compare(leaves[i], leaves[j]) < 0 })

	level := make([][]byte, len(leaves))
	for i, l := range leaves {
		h := sha256.Sum256(l)
		hh := sha256.Sum256(h[:])
		level[i] = hh[:]
	}
	for len(level) > 1 {
		if len(level)%2 == 1 {
			level = append(level, level[len(level)-1]) // duplicate last
		}
		var next [][]byte
		for i := 0; i < len(level); i += 2 {
			pair := append(level[i], level[i+1]...)
			h := sha256.Sum256(pair)
			hh := sha256.Sum256(h[:])
			next = append(next, hh[:])
		}
		level = next
	}
	root := make([]byte, 32)
	copy(root, level[0])
	return root, nil
}

//---------------------------------------------------------------------
// TLS config loader (TLS 1.3, X25519 Preferred)
//---------------------------------------------------------------------

func NewTLSConfig(certPath, keyPath string, requireClientCert bool) (*tls.Config, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS13,
		Certificates:             []tls.Certificate{cert},
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	if requireClientCert {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(certPEM) {
			return nil, errors.New("failed to append client cert to pool")
		}
		cfg.ClientCAs = pool
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return cfg, nil
}

// CertFingerprint returns the SHA-256 fingerprint of a PEM encoded certificate.
func CertFingerprint(certPath string) ([]byte, error) {
	pemData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}
	sum := sha256.Sum256(block.Bytes)
	fp := make([]byte, len(sum))
	copy(fp, sum[:])
	return fp, nil
}

// NewZeroTrustTLSConfig constructs a TLS 1.3 config with certificate pinning and
// optional mutual TLS.
func NewZeroTrustTLSConfig(certPath, keyPath, caPath string, pinnedFingerprint []byte) (*tls.Config, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	cfg := &tls.Config{
		MinVersion:             tls.VersionTLS13,
		MaxVersion:             tls.VersionTLS13,
		Certificates:           []tls.Certificate{cert},
		CurvePreferences:       []tls.CurveID{tls.X25519, tls.CurveP256},
		SessionTicketsDisabled: true,
	}

	if caPath != "" {
		caPEM, err := os.ReadFile(caPath)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caPEM) {
			return nil, errors.New("failed to load CA certificate")
		}
		cfg.ClientCAs = pool
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
	}

	if len(pinnedFingerprint) > 0 {
		fp := make([]byte, len(pinnedFingerprint))
		copy(fp, pinnedFingerprint)
		cfg.VerifyPeerCertificate = func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			if len(rawCerts) == 0 {
				return errors.New("no peer certificate provided")
			}
			hash := sha256.Sum256(rawCerts[0])
			if subtle.ConstantTimeCompare(hash[:], fp) != 1 {
				return fmt.Errorf("unexpected peer certificate fingerprint")
			}
			return nil
		}
	}

	return cfg, nil
}

// ---------------------------------------------------------------------
// Audit Trail & Predictive Security
// ---------------------------------------------------------------------

// AuditEvent represents a single immutable audit log entry.
type AuditEvent struct {
	Timestamp int64             `json:"ts"`
	Event     string            `json:"evt"`
	Meta      map[string]string `json:"meta,omitempty"`
	Hash      []byte            `json:"hash"`
}

// AuditTrail manages write-once audit logs with optional ledger anchoring.
type AuditTrail struct {
	mu     sync.Mutex
	file   *os.File
	ledger *Ledger
}

// NewAuditTrail creates or opens an append-only log file. If ledger is non-nil
// each entry hash is also stored on-chain for tamper evidence.
func NewAuditTrail(path string, ledger *Ledger) (*AuditTrail, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	return &AuditTrail{file: f, ledger: ledger}, nil
}

// Log writes an audit entry to disk and records its hash in the ledger.
func (a *AuditTrail) Log(event string, meta map[string]string) error {
	if a == nil || a.file == nil {
		return errors.New("audit trail not initialised")
	}
	ev := AuditEvent{Timestamp: time.Now().Unix(), Event: event, Meta: meta}
	raw, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	h := sha256.Sum256(raw)
	ev.Hash = h[:]
	blob, _ := json.Marshal(ev)
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, err := a.file.Write(append(blob, '\n')); err != nil {
		return err
	}
	if a.ledger != nil {
		key := append([]byte("audit:"), h[:]...)
		if err := a.ledger.SetState(key, h[:]); err != nil {
			return err
		}
	}
	return nil
}

// Report reads all audit entries from the log file.
func (a *AuditTrail) Report() ([]AuditEvent, error) {
	if a == nil || a.file == nil {
		return nil, errors.New("audit trail not initialised")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, err := a.file.Seek(0, 0); err != nil {
		return nil, err
	}
	var out []AuditEvent
	sc := bufio.NewScanner(a.file)
	for sc.Scan() {
		var ev AuditEvent
		if err := json.Unmarshal(sc.Bytes(), &ev); err == nil {
			out = append(out, ev)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Archive copies the current audit log to dest and writes a sha256 manifest.
// If dest is a directory, a timestamped file will be created inside it.
// The returned checksum is the hex-encoded SHA-256 of the log contents.
func (a *AuditTrail) Archive(dest string) (string, string, error) {
	if a == nil || a.file == nil {
		return "", "", errors.New("audit trail not initialised")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.file.Sync(); err != nil {
		return "", "", err
	}
	if _, err := a.file.Seek(0, 0); err != nil {
		return "", "", err
	}
	data, err := io.ReadAll(a.file)
	if err != nil {
		return "", "", err
	}
	if fi, err := os.Stat(dest); err == nil && fi.IsDir() {
		dest = filepath.Join(dest, fmt.Sprintf("audit_%d.log", time.Now().Unix()))
	}
	if err := os.WriteFile(dest, data, 0o600); err != nil {
		return "", "", err
	}
	sum := sha256.Sum256(data)
	checksum := fmt.Sprintf("%x", sum[:])
	manifest := fmt.Sprintf("%s  %s\n", checksum, filepath.Base(dest))
	if err := os.WriteFile(dest+".sha256", []byte(manifest), 0o600); err != nil {
		return "", "", err
	}
	return dest, checksum, nil
}

// Close closes the underlying log file.
func (a *AuditTrail) Close() error {
	if a == nil || a.file == nil {
		return nil
	}
	return a.file.Close()
}

// AnomalyDetector calculates streaming mean/variance for z-score detection.
type AnomalyDetector struct {
	mu    sync.RWMutex
	mean  float64
	m2    float64
	count int
}

// NewAnomalyDetector returns a new detector.
func NewAnomalyDetector() *AnomalyDetector { return &AnomalyDetector{} }

// Update incorporates a new observation.
func (ad *AnomalyDetector) Update(v float64) {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	ad.count++
	delta := v - ad.mean
	ad.mean += delta / float64(ad.count)
	ad.m2 += delta * (v - ad.mean)
}

// Score returns the absolute z-score for a value. If insufficient data is
// available the score is zero.
func (ad *AnomalyDetector) Score(v float64) float64 {
	ad.mu.RLock()
	mean, m2, n := ad.mean, ad.m2, ad.count
	ad.mu.RUnlock()
	if n < 2 {
		return 0
	}
	variance := m2 / float64(n-1)
	if variance == 0 {
		if v == mean {
			return 0
		}
		return math.Inf(1)
	}
	return math.Abs((v - mean) / math.Sqrt(variance))
}

// PredictRisk returns a moving average of the last window values, useful for
// simple trend-based security scoring.
func PredictRisk(values []float64, window int) float64 {
	if len(values) == 0 {
		return 0
	}
	if window <= 0 || window > len(values) {
		window = len(values)
	}
	sum := 0.0
	for _, v := range values[len(values)-window:] {
		sum += v
	}
	return sum / float64(window)
}

// ---------------------------------------------------------------------
// Quantum-Resistant Cryptography (Dilithium3)
// ---------------------------------------------------------------------

// DilithiumKeypair generates a Dilithium3 key pair.
func DilithiumKeypair() (pub, priv []byte, err error) {
	pk, sk, err := mode3.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return pk.Bytes(), sk.Bytes(), nil
}

// DilithiumSign signs msg with a packed Dilithium3 private key.
func DilithiumSign(priv, msg []byte) ([]byte, error) {
	var sk mode3.PrivateKey
	if err := sk.UnmarshalBinary(priv); err != nil {
		return nil, err
	}
	return sk.Sign(rand.Reader, msg, crypto.Hash(0))
}

// DilithiumVerify verifies a signature produced by DilithiumSign.
func DilithiumVerify(pub, msg, sig []byte) (bool, error) {
	var pk mode3.PublicKey
	if err := pk.UnmarshalBinary(pub); err != nil {
		return false, err
	}
	return mode3.Verify(&pk, msg, sig), nil
}
