package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ContentNode provides specialised handling for large encrypted content.
type ContentNode struct {
	*Node
	store map[string][]byte
}

// NewContentNode creates a network node capable of handling large content.
func NewContentNode(cfg Config) (*ContentNode, error) {
	n, err := NewNode(cfg)
	if err != nil {
		return nil, err
	}
	return &ContentNode{Node: n, store: make(map[string][]byte)}, nil
}


// encryptContent applies CFB encryption to the provided data using the key.
func encryptContent(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	b := make([]byte, aes.BlockSize+len(data))
	iv := b[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(b[aes.BlockSize:], data)
	return b, nil
}



// decryptContent reverses CFB encryption performed by encryptContent.
func decryptContent(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(data) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(data, data)
	return data, nil
}

// StoreContent encrypts and pins data returning its CID.
func (c *ContentNode) StoreContent(data, key []byte) (string, error) {
	enc, err := encryptContent(data, key)
	if err != nil {
		return "", err
	}
	cid, err := Pin(enc)
	if err != nil {
		return "", err
	}
	c.store[cid] = enc
	meta := ContentMeta{CID: cid, Size: uint64(len(enc)), Uploaded: time.Now().UTC()}
	raw, _ := json.Marshal(meta)
	if err := CurrentStore().Set([]byte("content:meta:"+cid), raw); err != nil {
		return cid, err
	}
	return cid, nil
}

// RetrieveContent fetches and decrypts content by CID.
func (c *ContentNode) RetrieveContent(cid string, key []byte) ([]byte, error) {
	data, ok := c.store[cid]
	if !ok {
		var err error
		data, err = Retrieve(cid)
		if err != nil {
			return nil, err
		}
	}
	return decryptContent(data, key)

}

// ListContent enumerates pinned content metadata.
func (c *ContentNode) ListContent() ([]ContentMeta, error) {
	it := CurrentStore().Iterator([]byte("content:meta:"), nil)
	defer it.Close()
	var list []ContentMeta
	for it.Next() {
		var m ContentMeta
		if err := json.Unmarshal(it.Value(), &m); err == nil {
			list = append(list, m)
		}
	}
	return list, nil
}
