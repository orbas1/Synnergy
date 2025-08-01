package core

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/rlp"
)

// HorizontalPartition splits data into fixed-size chunks. The last chunk
// may be shorter if the input length is not a multiple of size. A size of
// zero results in an empty slice.
func HorizontalPartition(data []byte, size int) [][]byte {
	if size <= 0 {
		return nil
	}
	var out [][]byte
	for len(data) > 0 {
		n := size
		if len(data) < size {
			n = len(data)
		}
		out = append(out, append([]byte(nil), data[:n]...))
		data = data[n:]
	}
	return out
}

// CompressData returns the gzip compressed form of the input slice.
func CompressData(in []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(in); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecompressData reverses CompressData.
func DecompressData(in []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	var out bytes.Buffer
	if _, err := io.Copy(&out, zr); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// PartitionAndCompress splits data into chunks, compressing each chunk
// individually.
func PartitionAndCompress(data []byte, size int) ([][]byte, error) {
	parts := HorizontalPartition(data, size)
	out := make([][]byte, len(parts))
	for i, p := range parts {
		c, err := CompressData(p)
		if err != nil {
			return nil, fmt.Errorf("compress chunk %d: %w", i, err)
		}
		out[i] = c
	}
	return out, nil
}

// DecompressAndCombine reverses PartitionAndCompress and concatenates
// the original data.
func DecompressAndCombine(parts [][]byte) ([]byte, error) {
	var buf bytes.Buffer
	for i, p := range parts {
		dec, err := DecompressData(p)
		if err != nil {
			return nil, fmt.Errorf("decompress chunk %d: %w", i, err)
		}
		buf.Write(dec)
	}
	return buf.Bytes(), nil
}

// StoreCompressedBlock serialises a block, compresses it in chunks and
// stores the segments in the ledger state under keys "compblk:<height>:<idx>".
func (l *Ledger) StoreCompressedBlock(b *Block, size int) error {
	enc, err := rlp.EncodeToBytes(b)
	if err != nil {
		return err
	}
	parts, err := PartitionAndCompress(enc, size)
	if err != nil {
		return err
	}
	for i, p := range parts {
		key := fmt.Sprintf("compblk:%d:%d", b.Header.Height, i)
		l.SetState([]byte(key), p)
	}
	return nil
}

// LoadCompressedBlock retrieves a compressed block from state and
// reconstructs it.
func (l *Ledger) LoadCompressedBlock(height uint64) (*Block, error) {
	var parts [][]byte
	for i := 0; ; i++ {
		key := fmt.Sprintf("compblk:%d:%d", height, i)
		val, err := l.GetState([]byte(key))
		if err != nil {
			break
		}
		parts = append(parts, val)
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("compressed block not found")
	}
	blob, err := DecompressAndCombine(parts)
	if err != nil {
		return nil, err
	}
	var blk Block
	if err := rlp.DecodeBytes(blob, &blk); err != nil {
		return nil, err
	}
	return &blk, nil
}
