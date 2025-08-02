package core

// ai_secure_storage.go - helpers for encrypted model parameter and dataset storage.

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// StoreModelParams encrypts and stores model parameters for the given hash.
func (ai *AIEngine) StoreModelParams(hash [32]byte, params []byte) error {
	if ai.encKey == nil {
		return fmt.Errorf("encryption key not initialised")
	}
	ct, err := encrypt(ai.encKey, params)
	if err != nil {
		return err
	}
	key := append([]byte("ai:modelparams:"), hash[:]...)
	return ai.led.SetState(key, ct)
}

// FetchModelParams retrieves and decrypts model parameters.
func (ai *AIEngine) FetchModelParams(hash [32]byte) ([]byte, error) {
	if ai.encKey == nil {
		return nil, fmt.Errorf("encryption key not initialised")
	}
	key := append([]byte("ai:modelparams:"), hash[:]...)
	raw, err := ai.led.GetState(key)
	if err != nil || raw == nil {
		return nil, fmt.Errorf("params not found: %w", err)
	}
	return decrypt(ai.encKey, raw)
}

// StoreDataset encrypts and persists training data referenced by ID.
func (ai *AIEngine) StoreDataset(id string, data []byte) error {
	if ai.encKey == nil {
		return fmt.Errorf("encryption key not initialised")
	}
	ct, err := encrypt(ai.encKey, data)
	if err != nil {
		return err
	}
	key := []byte("ai:dataset:" + id)
	return ai.led.SetState(key, ct)
}

// FetchDataset loads and decrypts a dataset by ID.
func (ai *AIEngine) FetchDataset(id string) ([]byte, error) {
	if ai.encKey == nil {
		return nil, fmt.Errorf("encryption key not initialised")
	}
	key := []byte("ai:dataset:" + id)
	raw, err := ai.led.GetState(key)
	if err != nil || raw == nil {
		return nil, fmt.Errorf("dataset not found: %w", err)
	}
	return decrypt(ai.encKey, raw)
}

func encrypt(key, plain []byte) ([]byte, error) {
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
	return gcm.Seal(nonce, nonce, plain, nil), nil
}

func decrypt(key, cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(cipherText) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce := cipherText[:gcm.NonceSize()]
	data := cipherText[gcm.NonceSize():]
	return gcm.Open(nil, nonce, data, nil)
}
