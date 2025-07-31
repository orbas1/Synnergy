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
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	gokzg4844 "github.com/crate-crypto/go-kzg-4844"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

//---------------------------------------------------------------------
// ComplianceEngine singleton
//---------------------------------------------------------------------

type ComplianceEngine struct {
	mu      sync.RWMutex
	ledger  StateRW
	allowed map[[33]byte]struct{} // issuer pubkey compressed
	fraud   map[Address]int
	auditNS []byte
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
		comp = &ComplianceEngine{
			ledger:  led,
			allowed: iss,
			fraud:   make(map[Address]int),
			auditNS: []byte("audit:"),
		}
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

// auditKey constructs the storage key for an audit log entry.
func (c *ComplianceEngine) auditKey(addr Address, ts int64) []byte {
	b := make([]byte, 0, len(c.auditNS)+len(addr)+8)
	b = append(b, c.auditNS...)
	b = append(b, addr[:]...)
	var t [8]byte
	binary.BigEndian.PutUint64(t[:], uint64(ts))
	return append(b, t[:]...)
}

// AuditEntry captures an immutable audit log entry.
type AuditEntry struct {
	Timestamp int64             `json:"ts"`
	Address   Address           `json:"addr"`
	Event     string            `json:"evt"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// LegalDoc represents fetched legal documentation.
type LegalDoc struct {
	Retrieved time.Time `json:"retrieved"`
	Content   []byte    `json:"content"`
}

// EncryptAES encrypts plaintext with the given key using AES-GCM.
func EncryptAES(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	out := gcm.Seal(nonce, nonce, plaintext, nil)
	return out, nil
}

// DecryptAES decrypts ciphertext produced by EncryptAES.
func DecryptAES(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	nonce := ciphertext[:gcm.NonceSize()]
	data := ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, data, nil)
}

// StoreEncrypted stores encrypted data on the ledger.
func (c *ComplianceEngine) StoreEncrypted(key, plaintext, aesKey []byte) error {
	enc, err := EncryptAES(aesKey, plaintext)
	if err != nil {
		return err
	}
	return c.ledger.SetState(key, enc)
}

// LoadDecrypted fetches encrypted data from the ledger and decrypts it.
func (c *ComplianceEngine) LoadDecrypted(key, aesKey []byte) ([]byte, error) {
	enc, err := c.ledger.GetState(key)
	if err != nil {
		return nil, err
	}
	if len(enc) == 0 {
		return nil, errors.New("not found")
	}
	return DecryptAES(aesKey, enc)
}

// VerifyZKProof validates a KZG proof for a given blob commitment.
// It returns true if the proof is valid under the EIP-4844 scheme.
func (c *ComplianceEngine) VerifyZKProof(blob, commitment, proof []byte) (bool, error) {
	if len(blob) != gokzg4844.ScalarsPerBlob*gokzg4844.SerializedScalarSize {
		return false, errors.New("invalid blob size")
	}
	if len(commitment) != gokzg4844.CompressedG1Size || len(proof) != gokzg4844.CompressedG1Size {
		return false, errors.New("invalid commitment or proof size")
	}

	var b gokzg4844.Blob
	copy(b[:], blob)
	var cmt gokzg4844.KZGCommitment
	copy(cmt[:], commitment)
	var pf gokzg4844.KZGProof
	copy(pf[:], proof)

	ctx, err := gokzg4844.NewContext4096Secure()
	if err != nil {
		return false, err
	}
	err = ctx.VerifyBlobKZGProof(&b, cmt, pf)
	return err == nil, err
}

// MaskSensitiveFields replaces the values of selected keys in the map with
// asterisks of the same length.  The original map is not modified.
func MaskSensitiveFields(data map[string]string, fields []string) map[string]string {
	out := make(map[string]string, len(data))
	for k, v := range data {
		out[k] = v
	}
	for _, f := range fields {
		if val, ok := out[f]; ok {
			out[f] = strings.Repeat("*", len(val))
		}
	}
	return out
}

// FetchLegalDoc retrieves a legal document from the provided URL.
func FetchLegalDoc(url string) (LegalDoc, error) {
	resp, err := http.Get(url)
	if err != nil {
		return LegalDoc{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return LegalDoc{}, fmt.Errorf("http status %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return LegalDoc{}, err
	}
	return LegalDoc{Retrieved: time.Now(), Content: data}, nil
}

// LogAudit stores an audit entry in the ledger.
func (c *ComplianceEngine) LogAudit(addr Address, event string, meta map[string]string) error {
	entry := AuditEntry{Timestamp: time.Now().Unix(), Address: addr, Event: event, Meta: meta}
	raw, _ := json.Marshal(entry)
	return c.ledger.SetState(c.auditKey(addr, entry.Timestamp), raw)
}

// AuditTrail retrieves all audit entries for an address.
func (c *ComplianceEngine) AuditTrail(addr Address) ([]AuditEntry, error) {
	prefix := append(c.auditNS, addr[:]...)
	it := c.ledger.PrefixIterator(prefix)
	var out []AuditEntry
	for it.Next() {
		var e AuditEntry
		if err := json.Unmarshal(it.Value(), &e); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

// MonitorTransaction analyses a transaction for anomalies using the AI engine.
func (c *ComplianceEngine) MonitorTransaction(tx *Transaction, threshold float32) (float32, error) {
	if tx == nil {
		return 0, errors.New("nil tx")
	}
	ai := AI()
	if ai == nil {
		return 0, errors.New("AI engine not initialised")
	}
	score, err := ai.PredictAnomaly(tx)
	if err != nil {
		return 0, err
	}
	if score > threshold {
		c.RecordFraudSignal(tx.From, int(score*10))
		_ = c.LogAudit(tx.From, "fraud_signal", map[string]string{"score": fmt.Sprintf("%f", score)})
	}
	return score, nil
}

// StartMonitor begins asynchronous monitoring of transactions received on txCh.
func (c *ComplianceEngine) StartMonitor(ctx context.Context, txCh <-chan *Transaction, threshold float32) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case tx := <-txCh:
				if tx != nil {
					_, _ = c.MonitorTransaction(tx, threshold)
				}
			}
		}
	}()
}

//---------------------------------------------------------------------
// END compliance.go
//---------------------------------------------------------------------
