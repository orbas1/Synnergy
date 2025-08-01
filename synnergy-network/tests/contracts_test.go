package core_test

import (
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	core "synnergy-network/core"
	"testing"
)

//------------------------------------------------------------
// Mock VM implementing Execute
//------------------------------------------------------------

type mockVM struct {
	returnData []byte
	err        error
	lastGas    uint64
	lastTxHash [32]byte
}

func (m *mockVM) Execute(code []byte, ctx *VMContext) (*Receipt, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.lastGas = ctx.GasLimit
	m.lastTxHash = ctx.TxHash
	return &Receipt{Status: true, ReturnData: m.returnData}, nil
}

//------------------------------------------------------------
// Helper to make a temp wasm file
//------------------------------------------------------------

func writeTempWasm(t *testing.T, data []byte) string {
	tmp, err := os.CreateTemp(t.TempDir(), "*.wasm")
	if err != nil {
		t.Fatalf("temp: %v", err)
	}
	if _, err = tmp.Write(data); err != nil {
		t.Fatalf("write: %v", err)
	}
	tmp.Close()
	return tmp.Name()
}

//------------------------------------------------------------
// Test CompileWASM
//------------------------------------------------------------

func TestCompileWASM(t *testing.T) {
	wasmBytes := []byte{0x00, 0x61, 0x73, 0x6d} // \0asm header
	wasmPath := writeTempWasm(t, wasmBytes)

	got, h, err := CompileWASM(wasmPath, t.TempDir())
	if err != nil {
		t.Fatalf("compile err %v", err)
	}
	if !reflect.DeepEqual(got, wasmBytes) {
		t.Fatalf("wasm bytes mismatch")
	}
	expHash := sha256.Sum256(wasmBytes)
	if h != expHash {
		t.Fatalf("hash mismatch")
	}

	// unsupported extension
	badPath := filepath.Join(t.TempDir(), "bad.txt")
	os.WriteFile(badPath, []byte("x"), 0o600)
	if _, _, err := CompileWASM(badPath, t.TempDir()); err == nil {
		t.Fatalf("expected error on bad ext")
	}
}

//------------------------------------------------------------
// Test deriveContractAddress deterministic
//------------------------------------------------------------

func TestDeriveContractAddress(t *testing.T) {
	creator := Address{0xAA}
	code := []byte{1, 2, 3}
	addr1 := deriveContractAddress(creator, code)
	addr2 := deriveContractAddress(creator, code)
	if addr1 != addr2 {
		t.Fatalf("determinism fail")
	}
}

//------------------------------------------------------------
// Test Deploy and Ricardian retrieval
//------------------------------------------------------------

func TestDeployAndRicardian(t *testing.T) {
	led, _ := NewInMemory()
	InitContracts(nil, &mockVM{})

	code := []byte{0x00, 0x61, 0x73, 0x6d}
	ric := []byte(`{"title":"test"}`)
	addr := DeriveContractAddress(Address{0x01}, code)

	if err := reg.Deploy(addr, code, ric, 50000); err != nil {
		t.Fatalf("deploy err %v", err)
	}

	// deploying again should fail
	if err := reg.Deploy(addr, code, nil, 0); err == nil {
		t.Fatalf("expected duplicate deploy error")
	}

	gotRic, err := reg.Ricardian(addr)
	if err != nil || string(gotRic) != string(ric) {
		t.Fatalf("ricardian retrieval failed: %v", err)
	}

	all := reg.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 contract, got %d", len(all))
	}
}

//------------------------------------------------------------
// Test Invoke success & error paths
//------------------------------------------------------------

func TestInvoke(t *testing.T) {
	// init registry
	InitContracts(nil, nil) // zero init

	// setup mock VM
	mvm := &mockVM{returnData: []byte("ok")}
	reg.vm = mvm

	// create fake contract
	sc := &SmartContract{Bytecode: []byte{0x00}, GasLimit: 50000}
	cAddr := Address{0x01}
	reg.byAddr[cAddr] = sc

	// pass larger gas should clamp
	ret, err := reg.Invoke(Address{0x02}, cAddr, "foo", nil, 1_000_000)
	if err != nil || string(ret) != "ok" {
		t.Fatalf("invoke fail %v", err)
	}
	if mvm.lastGas != sc.GasLimit {
		t.Fatalf("gas not clamped got %d want %d", mvm.lastGas, sc.GasLimit)
	}

	// contract not found
	if _, err := reg.Invoke(Address{0x02}, Address{0xFF}, "foo", nil, 0); err == nil {
		t.Fatalf("expected not found error")
	}

	// vm error propagation
	mvm.err = errors.New("vm boom")
	if _, err := reg.Invoke(Address{0x02}, cAddr, "foo", nil, 0); err == nil {
		t.Fatalf("expected vm error")
	}
}
