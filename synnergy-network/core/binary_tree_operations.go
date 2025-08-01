package core

import (
	"encoding/json"
	"fmt"
	"sync"
)

// BinaryTree provides a simple in-memory binary search tree that persists
// nodes in the ledger. Each operation records the state so that contracts and
// services may rely on deterministic storage.
// The tree is identified by name and scoped under the key prefix `bt:<name>`.
//
// This module is intentionally lightweight and integrates with the existing
// ledger interface and network broadcasting utilities. It can be invoked via
// the VM using dedicated opcodes and through the CLI using the `binarytree`
// command group.

type BinaryTree struct {
	name   string
	root   *btNode
	ledger *Ledger
	mu     sync.RWMutex
}

type btNode struct {
	Key   string  `json:"key"`
	Value []byte  `json:"value"`
	Left  *btNode `json:"left,omitempty"`
	Right *btNode `json:"right,omitempty"`
}

var (
	trees   = make(map[string]*BinaryTree)
	treesMu sync.RWMutex
)

// NewBinaryTree creates a new tree bound to the provided ledger. If a previous
// snapshot exists it is loaded automatically.
func NewBinaryTree(name string, led *Ledger) (*BinaryTree, error) {
	if led == nil {
		return nil, fmt.Errorf("ledger required")
	}
	bt := &BinaryTree{name: name, ledger: led}
	if err := bt.load(); err != nil && err.Error() != "state key not found" {
		return nil, err
	}
	treesMu.Lock()
	trees[name] = bt
	treesMu.Unlock()
	return bt, nil
}

// GetBinaryTree returns a previously created tree by name.
func GetBinaryTree(name string) *BinaryTree {
	treesMu.RLock()
	t := trees[name]
	treesMu.RUnlock()
	return t
}

// Insert adds or replaces a key in the tree and persists the change.
func (bt *BinaryTree) Insert(key string, value []byte) error {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	var err error
	bt.root, err = bt.insertRec(bt.root, key, value)
	if err != nil {
		return err
	}
	if err := bt.persist(); err != nil {
		return err
	}
	return bt.ledger.SetState(bt.nodeKey(key), value)
}

func (bt *BinaryTree) insertRec(n *btNode, key string, val []byte) (*btNode, error) {
	if n == nil {
		return &btNode{Key: key, Value: append([]byte(nil), val...)}, nil
	}
	switch {
	case key < n.Key:
		var err error
		n.Left, err = bt.insertRec(n.Left, key, val)
		return n, err
	case key > n.Key:
		var err error
		n.Right, err = bt.insertRec(n.Right, key, val)
		return n, err
	default:
		n.Value = append([]byte(nil), val...)
		return n, nil
	}
}

// Search returns the value stored under key. ok=false if not found.
func (bt *BinaryTree) Search(key string) ([]byte, bool) {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	n := bt.searchRec(bt.root, key)
	if n == nil {
		return nil, false
	}
	return append([]byte(nil), n.Value...), true
}

func (bt *BinaryTree) searchRec(n *btNode, key string) *btNode {
	if n == nil {
		return nil
	}
	switch {
	case key < n.Key:
		return bt.searchRec(n.Left, key)
	case key > n.Key:
		return bt.searchRec(n.Right, key)
	default:
		return n
	}
}

// Delete removes a key from the tree and persists the change.
func (bt *BinaryTree) Delete(key string) error {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	var deleted bool
	bt.root, deleted = bt.deleteRec(bt.root, key)
	if !deleted {
		return fmt.Errorf("key not found")
	}
	if err := bt.persist(); err != nil {
		return err
	}
	return bt.ledger.DeleteState(bt.nodeKey(key))
}

func (bt *BinaryTree) deleteRec(n *btNode, key string) (*btNode, bool) {
	if n == nil {
		return nil, false
	}
	switch {
	case key < n.Key:
		var del bool
		n.Left, del = bt.deleteRec(n.Left, key)
		return n, del
	case key > n.Key:
		var del bool
		n.Right, del = bt.deleteRec(n.Right, key)
		return n, del
	default:
		if n.Left == nil {
			return n.Right, true
		}
		if n.Right == nil {
			return n.Left, true
		}
		succ := n.Right
		for succ.Left != nil {
			succ = succ.Left
		}
		n.Key, n.Value = succ.Key, succ.Value
		var del bool
		n.Right, del = bt.deleteRec(n.Right, succ.Key)
		return n, del
	}
}

// InOrder returns all keys in sorted order.
func (bt *BinaryTree) InOrder() []string {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	var out []string
	bt.inOrderRec(bt.root, &out)
	return out
}

func (bt *BinaryTree) inOrderRec(n *btNode, out *[]string) {
	if n == nil {
		return
	}
	bt.inOrderRec(n.Left, out)
	*out = append(*out, n.Key)
	bt.inOrderRec(n.Right, out)
}

func (bt *BinaryTree) nodeKey(k string) []byte {
	return []byte(fmt.Sprintf("bt:%s:%s", bt.name, k))
}

func (bt *BinaryTree) persist() error {
	raw, err := json.Marshal(bt.root)
	if err != nil {
		return err
	}
	return bt.ledger.SetState([]byte("bt:"+bt.name+":root"), raw)
}

func (bt *BinaryTree) load() error {
	raw, err := bt.ledger.GetState([]byte("bt:" + bt.name + ":root"))
	if err != nil {
		return err
	}
	var root btNode
	if err := json.Unmarshal(raw, &root); err != nil {
		return err
	}
	bt.root = &root
	return nil
}

// -----------------------------------------------------------------------------
// Public helpers used by opcode dispatcher and CLI
// -----------------------------------------------------------------------------

// BinaryTreeNew initialises or loads a named tree.
func BinaryTreeNew(name string, led *Ledger) (*BinaryTree, error) {
	return NewBinaryTree(name, led)
}

// BinaryTreeInsert inserts a key/value pair into the named tree.
func BinaryTreeInsert(name, key string, value []byte) error {
	bt := GetBinaryTree(name)
	if bt == nil {
		return fmt.Errorf("tree %s not found", name)
	}
	return bt.Insert(key, value)
}

// BinaryTreeSearch retrieves the value for key from the named tree.
func BinaryTreeSearch(name, key string) ([]byte, bool, error) {
	bt := GetBinaryTree(name)
	if bt == nil {
		return nil, false, fmt.Errorf("tree %s not found", name)
	}
	val, ok := bt.Search(key)
	return val, ok, nil
}

// BinaryTreeDelete removes key from the named tree.
func BinaryTreeDelete(name, key string) error {
	bt := GetBinaryTree(name)
	if bt == nil {
		return fmt.Errorf("tree %s not found", name)
	}
	return bt.Delete(key)
}

// BinaryTreeInOrder returns all keys of the named tree in sorted order.
func BinaryTreeInOrder(name string) ([]string, error) {
	bt := GetBinaryTree(name)
	if bt == nil {
		return nil, fmt.Errorf("tree %s not found", name)
	}
	return bt.InOrder(), nil
}
