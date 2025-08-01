package core

import (
	"bytes"
	"crypto/sha256"
	"errors"
)

// BuildMerkleTree returns the level-by-level nodes of a Merkle tree built from
// the provided leaves. Each leaf is hashed using SHA-256. The last slice
// contains the single root hash.
func BuildMerkleTree(leaves [][]byte) ([][][32]byte, error) {
	if len(leaves) == 0 {
		return nil, errors.New("no leaves")
	}

	// first level: hashed leaves
	level := make([][32]byte, len(leaves))
	for i, l := range leaves {
		level[i] = sha256.Sum256(l)
	}

	tree := [][][32]byte{level}

	for len(level) > 1 {
		if len(level)%2 == 1 {
			level = append(level, level[len(level)-1])
		}
		next := make([][32]byte, len(level)/2)
		for i := 0; i < len(level); i += 2 {
			next[i/2] = sha256.Sum256(append(level[i][:], level[i+1][:]...))
		}
		tree = append(tree, next)
		level = next
	}

	return tree, nil
}

// MerkleProof returns a Merkle proof for the leaf at the given index along with
// the tree's root hash. The proof slice is ordered from leaf level upwards.
func MerkleProof(leaves [][]byte, index uint32) ([][]byte, [32]byte, error) {
	if len(leaves) == 0 {
		return nil, [32]byte{}, errors.New("no leaves")
	}
	if int(index) >= len(leaves) {
		return nil, [32]byte{}, errors.New("index out of range")
	}

	tree, err := BuildMerkleTree(leaves)
	if err != nil {
		return nil, [32]byte{}, err
	}

	proof := make([][]byte, 0, len(tree)-1)
	idx := int(index)
	for i := 0; i < len(tree)-1; i++ {
		level := tree[i]
		if idx%2 == 0 {
			proof = append(proof, level[idx+1][:])
		} else {
			proof = append(proof, level[idx-1][:])
		}
		idx /= 2
	}

	root := tree[len(tree)-1][0]
	return proof, root, nil
}

// VerifyMerklePath checks whether the supplied proof reconstructs the provided
// root for the given leaf and index. Proof hashes must be ordered from leaf
// upwards.
func VerifyMerklePath(root [32]byte, leaf []byte, proof [][]byte, index uint32) bool {
	h := sha256.Sum256(leaf)
	hash := h[:]
	for _, p := range proof {
		if index%2 == 0 {
			pair := append(hash, p...)
			sum := sha256.Sum256(pair)
			hash = sum[:]
		} else {
			pair := append(p, hash...)
			sum := sha256.Sum256(pair)
			hash = sum[:]
		}
		index /= 2
	}
	return bytes.Equal(hash, root[:])
}
