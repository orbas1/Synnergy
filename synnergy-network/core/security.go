// SPDX-License-Identifier: Apache-2.0
// Package core – shared security primitives for the Synnergy Network stack.
//
// Exposes:
//   • Sign / Verify      – Ed25519 (wallets) + BLS12-381 (validators).
//   • BLS aggregation    – multi-sig / threshold helpers.
//   • XChaCha20-Poly1305 – authenticated encryption.
//   • ComputeMerkleRoot – Bitcoin-style double-SHA256 Merkle tree.
//   • TLS loader         – hardened TLS 1.3 config for node-to-node gRPC.
//
// All crypto comes from Go 1.22 std-lib or herumi BLS (battle-tested).
package core

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	bls "github.com/herumi/bls-eth-go-binary/bls"
    
	"golang.org/x/crypto/chacha20poly1305"
)

//---------------------------------------------------------------------
// Package-level init – BLS curve setup
//---------------------------------------------------------------------

func init() {
	if err := bls.Init(0); err != nil {
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
	if len(key) != chacha20poly1305.KeySize {          // ← use KeySize
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
	if len(key) != chacha20poly1305.KeySize {          // ← use KeySize
		return nil, errors.New("key must be 32 bytes")
	}
	minLen := chacha20poly1305.NonceSizeX + chacha20poly1305.Overhead
	if len(blob) < minLen {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := blob[:chacha20poly1305.NonceSizeX], blob[minLen-len(blob):]
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
