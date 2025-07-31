package core

import (
    "crypto/sha256"
    "reflect"
    "testing"
	"encoding/hex"
)

// ------------------------------------------------------------
// Test AddBigInts helper
// ------------------------------------------------------------

func TestAddBigInts(t *testing.T) {
    cases := []struct {
        a, b string // hex strings for clarity
        want string
    }{
        {"01", "02", "03"},                  // 1 + 2
        {"ffffffff", "01", "100000000"},      // 2^32-1 +1
        {"", "", ""},                        // zero + zero
    }
    for _, tc := range cases {
        got := AddBigInts(fromHex(tc.a), fromHex(tc.b))
        if !reflect.DeepEqual(got, fromHex(tc.want)) {
            t.Fatalf("AddBigInts(%s,%s)=%x want %s", tc.a, tc.b, got, tc.want)
        }
    }
}

// ------------------------------------------------------------
// Test GasMeter behaviour
// ------------------------------------------------------------

func TestGasMeter(t *testing.T) {
    gm := NewGasMeter(5)
    if err := gm.Consume(PUSH); err != nil { // cost 2
        t.Fatalf("unexpected error: %v", err)
    }
    if err := gm.Consume(Opcode(0xFF)); err == nil { // unknown cost 100 -> OOG
        t.Fatal("expected out-of-gas error")
    }
}

// ------------------------------------------------------------
// SuperLightVM hash check
// ------------------------------------------------------------

func TestSuperLightVM(t *testing.T) {
    led, _ := NewInMemory()
    vm := &SuperLightVM{led}
    code := []byte{0x01, 0x02, 0x03}
    h := sha256.Sum256(code)
    ctx := &VMContext{TxHash: h}

    rec, err := vm.Execute(code, ctx)
    if err != nil || !rec.Status {
        t.Fatalf("expected success: %v %v", rec, err)
    }

    // mismatching hash
    badCtx := &VMContext{}
    rec, _ = vm.Execute(code, badCtx)
    if rec.Status {
        t.Fatal("expected failure due to hash mismatch")
    }
}

// ------------------------------------------------------------
// LightVM Execute normal + error cases
// ------------------------------------------------------------

func TestLightVM_Execute(t *testing.T) {
    led, _ := NewInMemory()

    // bytecode: PUSH 1, PUSH 2, ADD, RET => returns 3
    bc := []byte{
        byte(PUSH), 0x01, 0x01, // push 1
        byte(PUSH), 0x01, 0x02, // push 2
        byte(ADD),              // add
        byte(RET),              // return
    }

    vm := &LightVM{led: led, gas: NewGasMeter(100)}
    ctx := &VMContext{}
    rec, err := vm.Execute(bc, ctx)
    if err != nil || !rec.Status {
        t.Fatalf("unexpected exec err %v rec %+v", err, rec)
    }
    if len(rec.ReturnData) != 1 || rec.ReturnData[0] != 0x03 {
        t.Fatalf("expected return 3, got %x", rec.ReturnData)
    }

    // stack underflow: ADD with empty stack
    under := []byte{byte(ADD)}
    vm2 := &LightVM{led: led, gas: NewGasMeter(100)}
    rec, _ = vm2.Execute(under, ctx)
    if rec.Status {
        t.Fatal("expected failure on stack underflow")
    }

    // push length exceeds bounds
    bad := []byte{byte(PUSH), 0x05, 0x01}
    vm3 := &LightVM{led: led, gas: NewGasMeter(100)}
    rec, _ = vm3.Execute(bad, ctx)
    if rec.Status {
        t.Fatal("expected failure on push bounds")
    }
}

// ------------------------------------------------------------
// Helpers
// ------------------------------------------------------------

func fromHex(s string) []byte {
    if s == "" {
        return []byte{}
    }
    h, _ := hex.DecodeString(s)
    return h
}
