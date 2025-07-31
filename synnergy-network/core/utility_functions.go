package core

import (
	"math/big"
	"encoding/hex"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "errors"
    "crypto/sha256"
    "golang.org/x/crypto/sha3"
    "golang.org/x/crypto/ripemd160"
    "golang.org/x/crypto/blake2b"
    "fmt"
)

// Short returns a shortened hex version of the hash (e.g. first 4 + last 4).
func (h Hash) Short() string {
	hexStr := hex.EncodeToString(h[:])
	if len(hexStr) <= 8 {
		return hexStr
	}
	return hexStr[:4] + ".." + hexStr[len(hexStr)-4:]
}

// bytesToAddress converts a big-endian byte slice to a common.Address (20 bytes).
func BytesToAddress(b []byte) Address {
    var a Address
    if len(b) > len(a) {
        b = b[len(b)-len(a):]
    }
    copy(a[len(a)-len(b):], b)
    return a
}

func (s *Stack) Pop() *big.Int {
    if len(s.data) == 0 {
        panic("stack underflow")
    }
    idx := len(s.data) - 1
    raw := s.data[idx]
    s.data = s.data[:idx]
    // Assert it’s a *big.Int
    val, ok := raw.(*big.Int)
    if !ok {
        panic("stack element is not *big.Int")
    }
    return val
}


// Constants for 256-bit modular arithmetic.
var (
    two256  = new(big.Int).Lsh(big.NewInt(1), 256)
    mask256 = new(big.Int).Sub(two256, big.NewInt(1))
    two255  = new(big.Int).Lsh(big.NewInt(1), 255)
)

// opADD implements (a + b) mod 2^256.
func opADD(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    res := new(big.Int).Add(a, b)
    res.And(res, mask256)
    ctx.Stack.Push(res)
    return nil
}

// opMUL implements (a * b) mod 2^256.
func opMUL(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    res := new(big.Int).Mul(a, b)
    res.And(res, mask256)
    ctx.Stack.Push(res)
    return nil
}

// opSUB implements (b - a) mod 2^256.
func opSUB(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    res := new(big.Int).Sub(b, a)
    res.And(res, mask256)
    ctx.Stack.Push(res)
    return nil
}

// opDIV implements unsigned division: if a == 0 push 0, else floor(b / a).
func opDIV(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    if a.Sign() == 0 {
        ctx.Stack.Push(new(big.Int))
    } else {
        res := new(big.Int).Div(b, a)
        ctx.Stack.Push(res)
    }
    return nil
}

// opSDIV implements signed division: if a == 0 push 0, else trunc(b / a) with sign.
func opSDIV(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    if a.Sign() == 0 {
        ctx.Stack.Push(new(big.Int))
        return nil
    }
    bs := toSigned(b)
    as := toSigned(a)
    quot := new(big.Int).Div(bs, as)
    if quot.Sign() < 0 {
        quot.Add(quot, two256)
    }
    quot.And(quot, mask256)
    ctx.Stack.Push(quot)
    return nil
}

// opMOD implements unsigned modulo: if a == 0 push 0, else b mod a.
func opMOD(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    if a.Sign() == 0 {
        ctx.Stack.Push(new(big.Int))
    } else {
        res := new(big.Int).Mod(b, a)
        ctx.Stack.Push(res)
    }
    return nil
}

// opSMOD implements signed modulo per EVM: if a == 0 push 0, else sign(b) * (|b| mod |a|).
func opSMOD(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    as := toSigned(a)
    bs := toSigned(b)
    if as.Sign() == 0 {
        ctx.Stack.Push(new(big.Int))
        return nil
    }
    absA := new(big.Int).Abs(as)
    absB := new(big.Int).Abs(bs)
    r := new(big.Int).Mod(absB, absA)
    if bs.Sign() < 0 {
        r.Neg(r)
    }
    if r.Sign() < 0 {
        r.Add(r, two256)
    }
    r.And(r, mask256)
    ctx.Stack.Push(r)
    return nil
}

// opADDMOD implements (a + b) mod m, or 0 if m == 0.
func opADDMOD(ctx *VMContext) error {
    m := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    a := ctx.Stack.Pop()
    if m.Sign() == 0 {
        ctx.Stack.Push(new(big.Int))
        return nil
    }
    sum := new(big.Int).Add(a, b)
    res := new(big.Int).Mod(sum, m)
    ctx.Stack.Push(res)
    return nil
}

// opMULMOD implements (a * b) mod m, or 0 if m == 0.
func opMULMOD(ctx *VMContext) error {
    m := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    a := ctx.Stack.Pop()
    if m.Sign() == 0 {
        ctx.Stack.Push(new(big.Int))
        return nil
    }
    prod := new(big.Int).Mul(a, b)
    res := new(big.Int).Mod(prod, m)
    ctx.Stack.Push(res)
    return nil
}

// opEXP implements exponentiation modulo 2^256: a^b mod 2^256.
func opEXP(ctx *VMContext) error {
    exponent := ctx.Stack.Pop()
    base := ctx.Stack.Pop()
    res := new(big.Int).Exp(base, exponent, two256)
    ctx.Stack.Push(res)
    return nil
}

// opSIGNEXTEND extends the two's-complement signed value at byte index i.
func opSIGNEXTEND(ctx *VMContext) error {
    iBI := ctx.Stack.Pop()
    valBI := ctx.Stack.Pop()
    i := iBI.Uint64()
    val := new(big.Int).And(valBI, mask256)
    if i >= 32 {
        ctx.Stack.Push(val)
        return nil
    }
    bitPos := uint((i+1)*8 - 1)
    lowerMask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), bitPos+1), big.NewInt(1))
    bit := new(big.Int).Rsh(val, bitPos)
    bit.And(bit, big.NewInt(1))
    var res *big.Int
    if bit.Cmp(big.NewInt(1)) == 0 {
        inv := new(big.Int).Xor(lowerMask, mask256)
        res = new(big.Int).Or(val, inv)
    } else {
        res = new(big.Int).And(val, lowerMask)
    }
    res.And(res, mask256)
    ctx.Stack.Push(res)
    return nil
}

// opLT pushes 1 if b < a (unsigned), else 0.
func opLT(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    if b.Cmp(a) < 0 {
        ctx.Stack.Push(big.NewInt(1))
    } else {
        ctx.Stack.Push(big.NewInt(0))
    }
    return nil
}

// opGT pushes 1 if b > a (unsigned), else 0.
func opGT(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    if b.Cmp(a) > 0 {
        ctx.Stack.Push(big.NewInt(1))
    } else {
        ctx.Stack.Push(big.NewInt(0))
    }
    return nil
}

// opSLT pushes 1 if b < a (signed), else 0.
func opSLT(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    as := toSigned(a)
    bs := toSigned(b)
    if bs.Cmp(as) < 0 {
        ctx.Stack.Push(big.NewInt(1))
    } else {
        ctx.Stack.Push(big.NewInt(0))
    }
    return nil
}

// opSGT pushes 1 if b > a (signed), else 0.
func opSGT(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    as := toSigned(a)
    bs := toSigned(b)
    if bs.Cmp(as) > 0 {
        ctx.Stack.Push(big.NewInt(1))
    } else {
        ctx.Stack.Push(big.NewInt(0))
    }
    return nil
}

// opEQ pushes 1 if a == b, else 0.
func opEQ(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    if a.Cmp(b) == 0 {
        ctx.Stack.Push(big.NewInt(1))
    } else {
        ctx.Stack.Push(big.NewInt(0))
    }
    return nil
}

// opISZERO pushes 1 if a == 0, else 0.
func opISZERO(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    if a.Sign() == 0 {
        ctx.Stack.Push(big.NewInt(1))
    } else {
        ctx.Stack.Push(big.NewInt(0))
    }
    return nil
}

// opAND performs bitwise AND.
func opAND(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    res := new(big.Int).And(a, b)
    res.And(res, mask256)
    ctx.Stack.Push(res)
    return nil
}

// opOR performs bitwise OR.
func opOR(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    res := new(big.Int).Or(a, b)
    res.And(res, mask256)
    ctx.Stack.Push(res)
    return nil
}

// opXOR performs bitwise XOR.
func opXOR(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    b := ctx.Stack.Pop()
    res := new(big.Int).Xor(a, b)
    res.And(res, mask256)
    ctx.Stack.Push(res)
    return nil
}

// opNOT performs bitwise NOT.
func opNOT(ctx *VMContext) error {
    a := ctx.Stack.Pop()
    res := new(big.Int).Xor(a, mask256)
    ctx.Stack.Push(res)
    return nil
}

// opBYTE extracts the n-th byte (0 = most significant) from a 256-bit word.
func opBYTE(ctx *VMContext) error {
    nBI := ctx.Stack.Pop()
    valBI := ctx.Stack.Pop()
    n := nBI.Uint64()
    if n >= 32 {
        ctx.Stack.Push(big.NewInt(0))
        return nil
    }
    bytes := valBI.Bytes()
    padded := make([]byte, 32)
    copy(padded[32-len(bytes):], bytes)
    b := padded[n]
    ctx.Stack.Push(new(big.Int).SetUint64(uint64(b)))
    return nil
}

// opSHL implements logical left shift: (a << b) mod 2^256.
func opSHL(ctx *VMContext) error {
    shiftBI := ctx.Stack.Pop()
    valBI := ctx.Stack.Pop()
    shift := shiftBI.Uint64()
    if shift >= 256 {
        ctx.Stack.Push(big.NewInt(0))
        return nil
    }
    res := new(big.Int).Lsh(valBI, uint(shift))
    res.And(res, mask256)
    ctx.Stack.Push(res)
    return nil
}

// opSHR implements logical right shift (zero-fill).
func opSHR(ctx *VMContext) error {
    shiftBI := ctx.Stack.Pop()
    valBI := ctx.Stack.Pop()
    shift := shiftBI.Uint64()
    if shift >= 256 {
        ctx.Stack.Push(big.NewInt(0))
        return nil
    }
    val := new(big.Int).And(valBI, mask256)
    res := new(big.Int).Rsh(val, uint(shift))
    ctx.Stack.Push(res)
    return nil
}

// opSAR implements arithmetic right shift (sign-extend).
func opSAR(ctx *VMContext) error {
    shiftBI := ctx.Stack.Pop()
    valBI := ctx.Stack.Pop()
    shift := shiftBI.Uint64()
    val := new(big.Int).And(valBI, mask256)
    signed := toSigned(val)
    if shift >= 256 {
        if signed.Sign() < 0 {
            ctx.Stack.Push(new(big.Int).Set(mask256))
        } else {
            ctx.Stack.Push(big.NewInt(0))
        }
        return nil
    }
    res := new(big.Int).Rsh(signed, uint(shift))
    if res.Sign() < 0 {
        res.Add(res, two256)
    }
    res.And(res, mask256)
    ctx.Stack.Push(res)
    return nil
}

// toSigned converts an unsigned 256-bit big.Int to its signed two's-complement equivalent.
func toSigned(x *big.Int) *big.Int {
    if x.Cmp(two255) >= 0 {
        return new(big.Int).Sub(x, two256)
    }
    return new(big.Int).Set(x)
}

// ErrInvalidSignature is returned by opECRECOVER when the signature cannot be recovered.
var ErrInvalidSignature = errors.New("vm: invalid ecrecover signature")

// opECRECOVER pops 4 values (s, r, v, hash) from the stack (in that order),
// attempts to recover the public key from the ECDSA signature over secp256k1,
// converts it to an address, and pushes the 20-byte address (left-padded to 32 bytes).
// On failure, pushes 0.
func opECRECOVER(ctx *VMContext) error {
    // Stack layout (top-of-stack): [..., hash, v, r, s]
    sBI := ctx.Stack.Pop()
    rBI := ctx.Stack.Pop()
    vBI := ctx.Stack.Pop()
    hBI := ctx.Stack.Pop()

    hash := common.LeftPadBytes(hBI.Bytes(), 32)
    r := common.LeftPadBytes(rBI.Bytes(), 32)
    s := common.LeftPadBytes(sBI.Bytes(), 32)
    v := byte(vBI.Uint64())

    // Normalize v to {0,1}
    if v >= 27 {
        v -= 27
    }
    sig := append(append(r, s...), v)

    pubkey, err := crypto.SigToPub(hash, sig)
    if err != nil {
        // invalid signature: push 0
        ctx.Stack.Push(big.NewInt(0))
        return nil
    }
    addr := crypto.PubkeyToAddress(*pubkey)
    // push as 32-byte left-padded big.Int
    ctx.Stack.Push(new(big.Int).SetBytes(common.LeftPadBytes(addr.Bytes(), 32)))
    return nil
}

// bytesToAddress converts a big-endian byte slice to a common.Address (20 bytes).
func bytesToAddress(b []byte) common.Address {
    var a common.Address
    if len(b) > len(a) {
        b = b[len(b)-len(a):]
    }
    copy(a[len(a)-len(b):], b)
    return a
}

func opEXTCODESIZE(ctx *VMContext) error {
	addrBI := ctx.Stack.Pop()
	addr := Address(bytesToAddress(addrBI.Bytes())) // convert to your Address type
	code := ctx.State.GetCode(addr)
	ctx.Stack.Push(new(big.Int).SetUint64(uint64(len(code))))
	return nil
}

func opEXTCODECOPY(ctx *VMContext) error {
	length := ctx.Stack.Pop().Uint64()
	codeOffset := ctx.Stack.Pop().Uint64()
	memOffset := ctx.Stack.Pop().Uint64()
	addrBI := ctx.Stack.Pop()
	addr := Address(bytesToAddress(addrBI.Bytes())) // convert

	code := ctx.State.GetCode(addr)
	data := make([]byte, length)
	for i := uint64(0); i < length; i++ {
		if idx := codeOffset + i; idx < uint64(len(code)) {
			data[i] = code[idx]
		}
	}
	ctx.Memory.Write(memOffset, data)
	return nil
}

func opEXTCODEHASH(ctx *VMContext) error {
	addrBI := ctx.Stack.Pop()
	addr := Address(bytesToAddress(addrBI.Bytes())) // convert
	hash := ctx.State.GetCodeHash(addr)
	ctx.Stack.Push(new(big.Int).SetBytes(hash[:]))
	return nil
}


// opRETURNDATASIZE pushes the size of the return data buffer from the last external call.
func opRETURNDATASIZE(ctx *VMContext) error {
    size := uint64(len(ctx.LastReturnData))
    ctx.Stack.Push(new(big.Int).SetUint64(size))
    return nil
}

// opRETURNDATACOPY pops (memOffset, dataOffset, length) and copies that slice from the last return data into memory.
func opRETURNDATACOPY(ctx *VMContext) error {
    length := ctx.Stack.Pop().Uint64()
    dataOffset := ctx.Stack.Pop().Uint64()
    memOffset := ctx.Stack.Pop().Uint64()

    ret := ctx.LastReturnData
    data := make([]byte, length)
    for i := uint64(0); i < length; i++ {
        if idx := dataOffset + i; idx < uint64(len(ret)) {
            data[i] = ret[idx]
        }
    }
    ctx.Memory.Write(memOffset, data)
    return nil
}


// opMLOAD reads a 32-byte word from linear memory at the given offset.
func opMLOAD(ctx *VMContext) error {
    offset := ctx.Stack.Pop().Uint64()
    data := ctx.Memory.Read(offset, 32)
    ctx.Stack.Push(new(big.Int).SetBytes(data))
    return nil
}

// opMSTORE writes a 32-byte word into linear memory at the given offset.
func opMSTORE(ctx *VMContext) error {
    value := ctx.Stack.Pop()
    offset := ctx.Stack.Pop().Uint64()
    padded := common.LeftPadBytes(value.Bytes(), 32)
    ctx.Memory.Write(offset, padded)
    return nil
}

// opMSTORE8 writes the least-significant byte of the value into memory.
func opMSTORE8(ctx *VMContext) error {
    value := ctx.Stack.Pop().Uint64()
    offset := ctx.Stack.Pop().Uint64()
    b := byte(value & 0xff)
    ctx.Memory.Write(offset, []byte{b})
    return nil
}

// opCALLDATALOAD loads 32 bytes from the transaction args (zero-padded).
func opCALLDATALOAD(ctx *VMContext) error {
    offset := ctx.Stack.Pop().Uint64()
    var chunk [32]byte
    args := ctx.Args
    for i := uint64(0); i < 32; i++ {
        if idx := offset + i; idx < uint64(len(args)) {
            chunk[i] = args[idx]
        }
    }
    ctx.Stack.Push(new(big.Int).SetBytes(chunk[:]))
    return nil
}

// opCALLDATASIZE pushes the length of the transaction args.
func opCALLDATASIZE(ctx *VMContext) error {
    ctx.Stack.Push(new(big.Int).SetUint64(uint64(len(ctx.Args))))
    return nil
}

// opCALLDATACOPY copies a slice of args into memory.
func opCALLDATACOPY(ctx *VMContext) error {
    length := ctx.Stack.Pop().Uint64()
    dataOffset := ctx.Stack.Pop().Uint64()
    memOffset := ctx.Stack.Pop().Uint64()
    data := make([]byte, length)
    for i := uint64(0); i < length; i++ {
        if idx := dataOffset + i; idx < uint64(len(ctx.Args)) {
            data[i] = ctx.Args[idx]
        }
    }
    ctx.Memory.Write(memOffset, data)
    return nil
}

// opCODESIZE pushes the size of the contract code loaded in the VMContext.
func opCODESIZE(ctx *VMContext) error {
    ctx.Stack.Push(new(big.Int).SetUint64(uint64(len(ctx.Code))))
    return nil
}

// opCODECOPY copies a slice of the contract code into memory.
func opCODECOPY(ctx *VMContext) error {
    length := ctx.Stack.Pop().Uint64()
    codeOffset := ctx.Stack.Pop().Uint64()
    memOffset := ctx.Stack.Pop().Uint64()
    data := make([]byte, length)
    for i := uint64(0); i < length; i++ {
        if idx := codeOffset + i; idx < uint64(len(ctx.Code)) {
            data[i] = ctx.Code[idx]
        }
    }
    ctx.Memory.Write(memOffset, data)
    return nil
}


// ErrInvalidJumpDest is returned when a JUMP or JUMPI target is not a valid JUMPDEST.
var ErrInvalidJumpDest = errors.New("vm: invalid jump destination")

// opJUMP pops a destination off the stack and unconditionally sets PC to that location,
// if it is a valid JUMPDEST.
func opJUMP(ctx *VMContext) error {
    dest := ctx.Stack.Pop().Uint64()
    if _, ok := ctx.JumpTable[dest]; !ok {
        return ErrInvalidJumpDest
    }
    ctx.PC = dest
    return nil
}

// opJUMPI pops (dest, condition) and sets PC to dest if condition != 0 and dest is valid.
func opJUMPI(ctx *VMContext) error {
    dest := ctx.Stack.Pop().Uint64()
    cond := ctx.Stack.Pop()
    if cond.Sign() != 0 {
        if _, ok := ctx.JumpTable[dest]; !ok {
            return ErrInvalidJumpDest
        }
        ctx.PC = dest
    }
    return nil
}

// opPC pushes the index of the currently executing opcode onto the stack.
// Adjusts PC-1 to get the current instruction index.
func opPC(ctx *VMContext) error {
    cur := ctx.PC
    if cur > 0 {
        cur--
    }
    ctx.Stack.Push(new(big.Int).SetUint64(cur))
    return nil
}

// opMSIZE pushes the current size of active memory (in bytes).
func opMSIZE(ctx *VMContext) error {
    size := ctx.Memory.Len()
    ctx.Stack.Push(new(big.Int).SetUint64(uint64(size)))
    return nil
}

// opGAS pushes the remaining gas available to the execution context.
func opGAS(ctx *VMContext) error {
    rem := ctx.GasMeter.Remaining()
    ctx.Stack.Push(new(big.Int).SetUint64(rem))
    return nil
}

// opJUMPDEST is a no-op marking a valid jump destination.
func opJUMPDEST(ctx *VMContext) error {
    return nil
}


// ErrStop signals normal termination of execution.
var ErrStop = errors.New("vm: stop execution")

// returnError implements a return with data payload.
type returnError struct{ Data []byte }
func (e *returnError) Error() string      { return "vm: return" }
func (e *returnError) ReturnData() []byte { return e.Data }

// revertError implements a revert with data payload.
type revertError struct{ Data []byte }
func (e *revertError) Error() string      { return "vm: revert" }
func (e *revertError) ReturnData() []byte { return e.Data }

// opSHA256 computes SHA-256 over [offset, offset+size) in memory.
func opSHA256(ctx *VMContext) error {
    size := ctx.Stack.Pop().Uint64()
    offset := ctx.Stack.Pop().Uint64()
    data := ctx.Memory.Read(offset, size)
    sum := sha256.Sum256(data)
    ctx.Stack.Push(new(big.Int).SetBytes(sum[:]))
    return nil
}

// opKECCAK256 computes Keccak-256 over memory slice.
func opKECCAK256(ctx *VMContext) error {
    size := ctx.Stack.Pop().Uint64()
    offset := ctx.Stack.Pop().Uint64()
    data := ctx.Memory.Read(offset, size)
    h := sha3.NewLegacyKeccak256()
    h.Write(data)
    ctx.Stack.Push(new(big.Int).SetBytes(h.Sum(nil)))
    return nil
}

// opRIPEMD160 computes RIPEMD160, pads to 32 bytes.
func opRIPEMD160(ctx *VMContext) error {
    size := ctx.Stack.Pop().Uint64()
    offset := ctx.Stack.Pop().Uint64()
    data := ctx.Memory.Read(offset, size)
    h := ripemd160.New()
    h.Write(data)
    sum := h.Sum(nil)
    padded := make([]byte, 32)
    copy(padded[12:], sum)
    ctx.Stack.Push(new(big.Int).SetBytes(padded))
    return nil
}

// opBLAKE2B256 computes BLAKE2b-256.
func opBLAKE2B256(ctx *VMContext) error {
    size := ctx.Stack.Pop().Uint64()
    offset := ctx.Stack.Pop().Uint64()
    data := ctx.Memory.Read(offset, size)
    sum := blake2b.Sum256(data)
    ctx.Stack.Push(new(big.Int).SetBytes(sum[:]))
    return nil
}

// Environment ops
func opADDRESS(ctx *VMContext) error {
    ctx.Stack.Push(new(big.Int).SetBytes(ctx.Contract.Bytes()))
    return nil
}
func opCALLER(ctx *VMContext) error {
    ctx.Stack.Push(new(big.Int).SetBytes(ctx.Caller.Bytes()))
    return nil
}
func opORIGIN(ctx *VMContext) error {
    ctx.Stack.Push(new(big.Int).SetBytes(ctx.TxOrigin.Bytes()))
    return nil
}
func opCALLVALUE(ctx *VMContext) error {
    ctx.Stack.Push(ctx.Value)
    return nil
}
func opGASPRICE(ctx *VMContext) error {
    ctx.Stack.Push(new(big.Int).SetUint64(ctx.GasPrice))
    return nil
}

// Block context ops
func opNUMBER(ctx *VMContext) error {
    ctx.Stack.Push(new(big.Int).SetUint64(ctx.Chain.BlockNumber()))
    return nil
}
func opTIMESTAMP(ctx *VMContext) error {
    ctx.Stack.Push(new(big.Int).SetUint64(ctx.Chain.Time()))
    return nil
}
func opDIFFICULTY(ctx *VMContext) error {
    ctx.Stack.Push(ctx.Chain.Difficulty())
    return nil
}
func opGASLIMIT(ctx *VMContext) error {
    ctx.Stack.Push(new(big.Int).SetUint64(ctx.Chain.GasLimit()))
    return nil
}
func opCHAINID(ctx *VMContext) error {
    ctx.Stack.Push(ctx.Chain.ChainID())
    return nil
}
func opBLOCKHASH(ctx *VMContext) error {
    n := ctx.Stack.Pop().Uint64()
    h := ctx.Chain.BlockHash(n)
    ctx.Stack.Push(new(big.Int).SetBytes(h.Bytes()))
    return nil
}

// Account ops
func opBALANCE(ctx *VMContext) error {
    addr := Address(common.BytesToAddress(ctx.Stack.Pop().Bytes())) // ✅ cast to custom Address
    bal := ctx.State.BalanceOf(addr)
    ctx.Stack.Push(new(big.Int).SetUint64(bal)) // ensure big.Int
    return nil
}


func opSELFBALANCE(ctx *VMContext) error {
    bal := ctx.State.BalanceOf(ctx.Contract)
    ctx.Stack.Push(bal)
    return nil
}

// Logging ops
func opLOG0(ctx *VMContext) error { return logN(ctx, 0) }
func opLOG1(ctx *VMContext) error { return logN(ctx, 1) }
func opLOG2(ctx *VMContext) error { return logN(ctx, 2) }
func opLOG3(ctx *VMContext) error { return logN(ctx, 3) }
func opLOG4(ctx *VMContext) error { return logN(ctx, 4) }
func logN(ctx *VMContext, n int) error {
    topics := make([]common.Hash, n)
    for i := n - 1; i >= 0; i-- {
        topics[i] = common.BigToHash(ctx.Stack.Pop())
    }
    size := ctx.Stack.Pop().Uint64()
    offset := ctx.Stack.Pop().Uint64()
    data := ctx.Memory.Read(offset, size)
    ctx.State.AddLog(&Log{Address: ctx.Contract, Topics: topics, Data: data})
    return nil
}




// Calls and creation
func opCREATE(ctx *VMContext) error {
    value := ctx.Stack.Pop()
    size := ctx.Stack.Pop().Uint64()
    offset := ctx.Stack.Pop().Uint64()
    gas := ctx.Stack.Pop().Uint64()
    code := ctx.Memory.Read(offset, size)
    addr, _, ok, _ := ctx.State.CreateContract(ctx.Contract, code, value, gas)
    if ok {
        ctx.Stack.Push(new(big.Int).SetBytes(addr.Bytes()))
    } else {
        ctx.Stack.Push(big.NewInt(0))
    }
    return nil
}
func opCALL(ctx *VMContext) error { return call(ctx, false) }
func opCALLCODE(ctx *VMContext) error { return call(ctx, true) }
// Delegate and static calls
func opDELEGATECALL(ctx *VMContext) error {
	to := Address(common.BytesToAddress(ctx.Stack.Pop().Bytes()))
	inOff := ctx.Stack.Pop().Uint64()
	inSz := ctx.Stack.Pop().Uint64()
	input := ctx.Memory.Read(inOff, inSz)
	value := ctx.Stack.Pop()
	gas := ctx.Stack.Pop().Uint64()

	return ctx.State.DelegateCall(ctx.Contract, to, input, value, gas)
}

func opSTATICCALL(ctx *VMContext) error {
    return callStatic(ctx)
}

func call(ctx *VMContext, codeOnly bool) error {
	outSz := ctx.Stack.Pop().Uint64()
	outOff := ctx.Stack.Pop().Uint64()
	inSz := ctx.Stack.Pop().Uint64()
	inOff := ctx.Stack.Pop().Uint64()
	value := ctx.Stack.Pop()
	to := BytesToAddress(ctx.Stack.Pop().Bytes())
	gas := ctx.Stack.Pop().Uint64()
	data := ctx.Memory.Read(inOff, inSz)

	var ret []byte
	var ok bool
	if codeOnly {
		ret, ok, _ = ctx.State.CallCode(ctx.Contract, to, data, value, gas)
	} else {
		ret, ok, _ = ctx.State.CallContract(ctx.Contract, to, data, value, gas)
	}

	if ok {
		ctx.Memory.Write(outOff, ret[:min(uint64(len(ret)), outSz)]) // ✅ safe truncate
		ctx.Stack.Push(big.NewInt(1))
	} else {
		ctx.Stack.Push(big.NewInt(0))
	}
	return nil
}


func callStatic(ctx *VMContext) error {
	outSz := ctx.Stack.Pop().Uint64()
	outOff := ctx.Stack.Pop().Uint64()
	inSz := ctx.Stack.Pop().Uint64()
	inOff := ctx.Stack.Pop().Uint64()
	to := BytesToAddress(ctx.Stack.Pop().Bytes())
	gas := ctx.Stack.Pop().Uint64()

	input := ctx.Memory.Read(inOff, inSz)
	ret, ok, _ := ctx.State.StaticCall(ctx.Contract, to, input, gas)

	if ok {
		ctx.Memory.Write(outOff, ret[:min(uint64(len(ret)), outSz)]) // ✅ truncate safely
		ctx.Stack.Push(big.NewInt(1))
	} else {
		ctx.Stack.Push(big.NewInt(0))
	}
	return nil
}



// Termination
func opRETURN(ctx *VMContext) error {
    sz := ctx.Stack.Pop().Uint64()
    off := ctx.Stack.Pop().Uint64()
    return &returnError{Data: ctx.Memory.Read(off, sz)}
}
func opREVERT(ctx *VMContext) error {
    sz := ctx.Stack.Pop().Uint64()
    off := ctx.Stack.Pop().Uint64()
    return &revertError{Data: ctx.Memory.Read(off, sz)}
}
func opSTOP(ctx *VMContext) error { return ErrStop }
func opSELFDESTRUCT(ctx *VMContext) error {
    ben := BytesToAddress(ctx.Stack.Pop().Bytes())
    ctx.State.SelfDestruct(ctx.Contract, ben)
    return ErrStop
}



type AssetKind int

const (
    AssetCoin AssetKind = iota
    AssetToken
)


type AssetRef struct {
    Kind    AssetKind // coin or token
    TokenID TokenID   // only required for tokens
}


var (
    CoinLedger  *Coin
    TokenLedger = make(map[TokenID]*BaseToken)
)


func Transfer(ctx *Context, asset AssetRef, from, to Address, amount uint64) error {
	switch asset.Kind {
	case AssetCoin:
		return ctx.State.Transfer(from, to, amount) // ✅ fixed

	case AssetToken:
		token, ok := TokenLedger[asset.TokenID]
		if !ok {
			return fmt.Errorf("token not found: %x", asset.TokenID)
		}
		return token.Transfer(from, to, amount) // ✅ fixed

	default:
		return fmt.Errorf("invalid asset kind")
	}
}

func Mint(ctx *Context, asset AssetRef, to Address, amount uint64) error {
	switch asset.Kind {
	case AssetCoin:
		return ctx.State.Mint(to, amount) // ✅ fixed

	case AssetToken:
		token, ok := TokenLedger[asset.TokenID]
		if !ok {
			return fmt.Errorf("token not found: %x", asset.TokenID)
		}
		return token.Mint(to, amount) // ✅ fixed

	default:
		return fmt.Errorf("invalid asset kind")
	}
}

func Burn(ctx *Context, asset AssetRef, from Address, amount uint64) error {
	switch asset.Kind {
	case AssetCoin:
		return ctx.State.Burn(from, amount) // ✅ fixed

	case AssetToken:
		token, ok := TokenLedger[asset.TokenID]
		if !ok {
			return fmt.Errorf("token not found: %x", asset.TokenID)
		}
		return token.Burn(from, amount) // ✅ fixed

	default:
		return fmt.Errorf("invalid asset kind")
	}
}
