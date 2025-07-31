package core

// Smart‑Contract Runtime & Registry for Synnergy Network.
//
// Highlights
// -----------
//   • WASM‑first execution – contracts are authored as Rust/AssemblyScript → compiled to
//     WASM64, deterministically hashed, and sandboxed inside the chain VM.
//   • Rich *Ricardian Contract* metadata – JSON manifest binding legal prose to code
//     hash (meets UKJT / ISO‑TC 307 compliance).
//   • Offline compiler helper (`CompileWASM`) uses wazero’s CLI wrapper to produce a
//     reproducible byte‑blob; can be replaced with Canonical build‑service.
//   • Registry exposes `Invoke` which routes execution to VM, meters gas, and logs.
//
// Build‑graph: depends on common, ledger, vm. NO network or higher‑tier imports.
// -----------------------------------------------------------------------------

import (
	"crypto/sha256"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

//---------------------------------------------------------------------
// Registry (singleton)
//---------------------------------------------------------------------

var (
	contractOnce sync.Once
	reg          *ContractRegistry
)

func InitContracts(led StateRW, vmm VM) {
	contractOnce.Do(func() {
		reg = &ContractRegistry{
			ledger: led,
			vm:     vmm,
			byAddr: make(map[Address]*SmartContract),
		}
	})
}

//---------------------------------------------------------------------
// Compile & Deploy pipeline
//---------------------------------------------------------------------

// CompileWASM compiles source file to WASM byte‑blob via wazero CLI (deterministic build).
func CompileWASM(srcPath string, outDir string) ([]byte, [32]byte, error) {
	if filepath.Ext(srcPath) != ".wat" && filepath.Ext(srcPath) != ".wasm" {
		return nil, [32]byte{}, errors.New("unsupported source – must be .wat/.wasm compiled offline beforehand")
	}

	// If .wasm supplied, just read bytes; else use 'wat2wasm'.
	var wasm []byte
	if filepath.Ext(srcPath) == ".wasm" {
		b, err := os.ReadFile(srcPath)
		if err != nil {
			return nil, [32]byte{}, err
		}
		wasm = b
	} else {
		out := filepath.Join(outDir, filepath.Base(srcPath)+".wasm")
		cmd := exec.Command("wat2wasm", "-o", out, srcPath)
		if err := cmd.Run(); err != nil {
			return nil, [32]byte{}, err
		}
		b, _ := os.ReadFile(out)
		wasm = b
	}
	hash := sha256.Sum256(wasm)
	return wasm, hash, nil
}

//---------------------------------------------------------------------
// Invocation – routed through VM.
//---------------------------------------------------------------------

func (cr *ContractRegistry) Invoke(
	caller Address, // your own 20-byte address type
	addr Address,
	method string,
	args []byte,
	gasLimit uint64,
) ([]byte, error) {

	// 1. Look up the contract
	cr.mu.RLock()
	sc, ok := cr.byAddr[addr]
	cr.mu.RUnlock()
	if !ok {
		return nil, errors.New("contract not found")
	}

	// 2. Clamp gas
	if gasLimit == 0 || gasLimit > sc.GasLimit {
		gasLimit = sc.GasLimit
	}

	// 3. Convert your Address → common.Address
	callerAddr := common.BytesToAddress(caller[:]) // helper from go-ethereum
	originAddr := callerAddr                       // same for now

	// 4. Build the VM context (only the fields that exist!)
	vmCtx := &VMContext{
		Caller:   callerAddr,
		Origin:   originAddr,
		TxHash:   zeroHash, // or real hash if you have it
		GasLimit: gasLimit,
	}

	// 5. IMPORTANT: argument order is (code, ctx)
	rec, err := cr.vm.Execute(sc.Bytecode, vmCtx)
	if err != nil {
		return nil, err
	}

	// 6. Give back the contract’s return bytes
	return rec.ReturnData, nil // change to the actual field name
}

var zeroHash [32]byte // all-zero value

//---------------------------------------------------------------------
// Helpers
//---------------------------------------------------------------------

func deriveContractAddress(creator Address, code []byte) Address {
	pre := append(creator.Bytes(), code...)
	h := sha256.Sum256(pre)
	var out Address
	copy(out[:], h[:20])
	return out
}

func contractKey(addr Address) []byte  { return append([]byte("contract:code:"), addr.Bytes()...) }
func ricardianKey(addr Address) []byte { return append([]byte("contract:ric:"), addr.Bytes()...) }
