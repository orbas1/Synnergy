package core_test

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	core "synnergy-network/core"
	"testing"
	"time"
)

//------------------------------------------------------------
// Helper to generate a self‑signed cert PEM pair for TLS tests
//------------------------------------------------------------

func genSelfSignedCert(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()
	priv, err := tls.GeneratePrivateKey(tls.RSAWithSHA256, 2048)
	if err != nil {
		t.Fatalf("priv gen: %v", err)
	}
	template := &x509.Certificate{SerialNumber: new(big.Int).SetInt64(1), NotBefore: time.Now(), NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, template, template, priv.Public(), priv)
	if err != nil {
		t.Fatalf("cert create: %v", err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("priv der: %v", err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})
	return
}

//------------------------------------------------------------
// Sign / Verify (Ed25519)
//------------------------------------------------------------

func TestSignVerify_Ed25519(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	msg := []byte("hello")
	sig, err := Sign(AlgoEd25519, priv, msg)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	ok, err := Verify(AlgoEd25519, pub, msg, sig)
	if err != nil || !ok {
		t.Fatalf("verify failed err=%v ok=%v", err, ok)
	}

	// corrupt sig
	sig[0] ^= 0xFF
	ok, _ = Verify(AlgoEd25519, pub, msg, sig)
	if ok {
		t.Fatalf("expected verify=false on corrupted sig")
	}
}

func TestSignVerify_Errors(t *testing.T) {
	_, err := Sign(99, nil, nil)
	if err == nil {
		t.Fatalf("expected unknown algo error")
	}

	_, err = Sign(AlgoEd25519, "notkey", nil)
	if err == nil {
		t.Fatalf("expected invalid key type error")
	}

	_, err = Verify(AlgoEd25519, "badpub", nil, nil)
	if err == nil {
		t.Fatalf("expected invalid pub error")
	}
}

//------------------------------------------------------------
// CombineShares – trivial threshold=1 success & failure path
//------------------------------------------------------------

func TestCombineShares(t *testing.T) {
	secret := make([]byte, 32)
	rand.Read(secret)
	shares := []Share{{Index: 1, Data: secret}}
	got, err := CombineShares(shares, 1)
	if err != nil {
		t.Fatalf("combine err %v", err)
	}
	if !bytes.Equal(got, secret) {
		t.Fatalf("combine mismatch")
	}

	if _, err := CombineShares(shares, 2); err == nil {
		t.Fatalf("expected not enough shares error")
	}
}

//------------------------------------------------------------
// Encrypt / Decrypt round‑trip & error paths
//------------------------------------------------------------

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)
	msg := []byte("secret data")
	aad := []byte("hdr")

	blob, err := Encrypt(key, msg, aad)
	if err != nil {
		t.Fatalf("enc err %v", err)
	}
	plain, err := Decrypt(key, blob, aad)
	if err != nil {
		t.Fatalf("dec err %v", err)
	}
	if !bytes.Equal(plain, msg) {
		t.Fatalf("decrypt mismatch")
	}

	// wrong key size
	if _, err := Encrypt(make([]byte, 16), msg, nil); err == nil {
		t.Fatalf("expected key size error")
	}

	// tamper ciphertext
	blob[len(blob)-1] ^= 0x01
	if _, err := Decrypt(key, blob, aad); err == nil {
		t.Fatalf("expected decrypt auth error")
	}
}

//------------------------------------------------------------
// ComputeMerkleRoot
//------------------------------------------------------------

func TestComputeMerkleRoot(t *testing.T) {
	if _, err := ComputeMerkleRoot(nil); err == nil {
		t.Fatalf("want error on empty leaves")
	}
	a := sha256.Sum256([]byte("a"))
	b := sha256.Sum256([]byte("b"))
	root1, _ := ComputeMerkleRoot([][]byte{a[:], b[:]})
	root2, _ := ComputeMerkleRoot([][]byte{b[:], a[:]}) // order changed, roots equal due to sort
	if !bytes.Equal(root1, root2) {
		t.Fatalf("root mismatch, sort failed")
	}
	if len(root1) != 32 {
		t.Fatalf("root len %d", len(root1))
	}
}

//------------------------------------------------------------
// TLS config loader minimal smoke
//------------------------------------------------------------

func TestNewTLSConfig(t *testing.T) {
	certPEM, keyPEM := genSelfSignedCert(t)

	certFile, _ := ioutil.TempFile(t.TempDir(), "cert.pem")
	keyFile, _ := ioutil.TempFile(t.TempDir(), "key.pem")
	os.WriteFile(certFile.Name(), certPEM, 0600)
	os.WriteFile(keyFile.Name(), keyPEM, 0600)

	cfg, err := NewTLSConfig(certFile.Name(), keyFile.Name(), false)
	if err != nil {
		t.Fatalf("tls cfg err %v", err)
	}
	if cfg.MinVersion != tls.VersionTLS13 {
		t.Fatalf("min version not TLS1.3")
	}
}
