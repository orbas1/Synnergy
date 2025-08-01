package core

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// CompressLedger returns the gzip-compressed JSON encoding of the provided ledger.
func CompressLedger(l *Ledger) ([]byte, error) {
	if l == nil {
		return nil, fmt.Errorf("nil ledger")
	}
	data, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecompressLedger reverses CompressLedger.
func DecompressLedger(data []byte) (*Ledger, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	var out bytes.Buffer
	if _, err := io.Copy(&out, gr); err != nil {
		return nil, err
	}
	var l Ledger
	if err := json.Unmarshal(out.Bytes(), &l); err != nil {
		return nil, err
	}
	return &l, nil
}

// SaveCompressedSnapshot writes the ledger snapshot compressed with gzip to the specified path.
func SaveCompressedSnapshot(l *Ledger, path string) error {
	data, err := CompressLedger(l)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// LoadCompressedSnapshot reads a gzip compressed snapshot from path and returns the ledger.
func LoadCompressedSnapshot(path string) (*Ledger, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return DecompressLedger(data)
}
