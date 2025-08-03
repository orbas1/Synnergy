package core_test

import (
	"errors"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/wasmerio/wasmer-go/wasmer"

	core "synnergy-network/core"
)

// TestHeavyVMInvokeWithReceipt compiles a sample contract, deploys it to the
// in-memory registry and verifies that logs are captured in the receipt.
func TestHeavyVMInvokeWithReceipt(t *testing.T) {
	watPath := filepath.Join("cmd", "smart_contracts", "examples", "log.wat")
	wasm, _, err := core.CompileWASM(watPath, t.TempDir())
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			t.Skip("wat2wasm not installed")
		}
		t.Fatalf("compile wasm: %v", err)
	}

	led, _ := core.NewInMemory()
	vm := core.NewHeavyVM(led, core.NewGasMeter(1_000_000), wasmer.NewEngine())
	core.InitContracts(led, vm)

	addr := core.DeriveContractAddress(core.AddressZero, wasm)
	if err := core.GetContractRegistry().Deploy(addr, wasm, nil, 1_000_000); err != nil {
		t.Fatalf("deploy contract: %v", err)
	}

	rec, err := core.GetContractRegistry().InvokeWithReceipt(core.AddressZero, addr, "", nil, 0)
	if err != nil || !rec.Status {
		t.Fatalf("invoke error: %v %+v", err, rec)
	}
	if len(rec.Logs) != 1 || string(rec.Logs[0].Data) != "hello" {
		t.Fatalf("unexpected logs: %+v", rec.Logs)
	}
}
