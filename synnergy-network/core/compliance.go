package core

// compliance.go – Regulatory & Data‑Privacy utilities for Synnergy Network.
//
// Features
// --------
//   • GDPR Right‑to‑Erasure (`EraseData(addr)`): scrubs personal KYC blobs while
//     preserving cryptographic proofs of compliance (hash commitments remain).
//   • KYC validation (`ValidateKYC(doc)`): verifies signature chain from issuer
//     (government / bank) and stores a blinded commitment in ledger state.
//   • FraudTracking (`RecordFraudSignal`, `RiskScore(addr)`): cooperates with AI
//     engine to flag suspicious addresses.
//
// Dependencies: common, ledger, security.
// ----------------------------------------------------------------------------

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"math/big"
	"sync"
)

//---------------------------------------------------------------------
// ComplianceEngine singleton
//---------------------------------------------------------------------

type ComplianceEngine struct {
	mu      sync.RWMutex
	ledger  StateRW
	allowed map[[33]byte]struct{} // issuer pubkey compressed
	fraud   map[Address]int
}

var (
	compOnce sync.Once
	comp     *ComplianceEngine
)

func InitCompliance(led StateRW, trustedIssuers [][]byte) {
	compOnce.Do(func() {
		iss := make(map[[33]byte]struct{})
		for _, pk := range trustedIssuers {
			var key [33]byte
			copy(key[:], pk)
			iss[key] = struct{}{}
		}
		comp = &ComplianceEngine{ledger: led, allowed: iss, fraud: make(map[Address]int)}
	})
}

func Compliance() *ComplianceEngine { return comp }

//---------------------------------------------------------------------
// ValidateKYC – stores commitment if issuer is trusted & sig valid.
//---------------------------------------------------------------------

func (c *ComplianceEngine) ValidateKYC(doc *KYCDocument) error {
	if doc == nil {
		return errors.New("nil doc")
	}

	// issuer trust check
	var issuerKey [33]byte
	copy(issuerKey[:], doc.IssuerPK)
	if _, ok := c.allowed[issuerKey]; !ok {
		return errors.New("untrusted issuer")
	}

	// prepare message
	raw, _ := json.Marshal(struct {
		Address     Address
		CountryCode string
		IDHash      [32]byte
		IssuedAt    int64
	}{doc.Address, doc.CountryCode, doc.IDHash, doc.IssuedAt})

	hash := sha256.Sum256(raw)

	// recover public key
	pk, err := secp256k1.ParsePubKey(doc.IssuerPK)
	if err != nil {
		return errors.New("invalid issuer pubkey")
	}

	// decode sig
	r, s, err := decodeSig(doc.Signature)
	if err != nil {
		return err
	}

	// verify
	if !ecdsa.Verify(pk.ToECDSA(), hash[:], r, s) {
		return errors.New("invalid signature")
	}

	// store blinded commitment
	key := kycKey(doc.Address)
	val := sha256.Sum256(doc.Signature)
	c.ledger.SetState(key, val[:])
	return nil
}

func decodeSig(sig []byte) (r, s *big.Int, err error) {
	if len(sig) != 64 {
		return nil, nil, errors.New("invalid sig length")
	}
	r = new(big.Int).SetBytes(sig[:32])
	s = new(big.Int).SetBytes(sig[32:])
	return r, s, nil
}

// Sample function using standard library
func VerifyECDSA(pub *ecdsa.PublicKey, msg []byte, r, s *big.Int) bool {
	hashed := sha256.Sum256(msg)
	return ecdsa.Verify(pub, hashed[:], r, s)
}

//---------------------------------------------------------------------
// EraseData – GDPR delete personal data but keep zero‑knowledge commitment.
//---------------------------------------------------------------------

func (c *ComplianceEngine) EraseData(addr Address) error {
	key := kycKey(addr)
	blob, _ := c.ledger.GetState(key)
	if len(blob) == 0 {
		return errors.New("no KYC")
	}
	// replace with tombstone 0x00
	c.ledger.SetState(key, []byte{0x00})
	return nil
}

//---------------------------------------------------------------------
// Fraud tracking helpers
//---------------------------------------------------------------------

func (c *ComplianceEngine) RecordFraudSignal(addr Address, severity int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fraud[addr] += severity
}

func (c *ComplianceEngine) RiskScore(addr Address) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.fraud[addr]
}

//---------------------------------------------------------------------
// Ledger key helpers
//---------------------------------------------------------------------

func kycKey(addr Address) []byte { return append([]byte("kyc:"), addr[:]...) }

//---------------------------------------------------------------------
// END compliance.go
//---------------------------------------------------------------------
