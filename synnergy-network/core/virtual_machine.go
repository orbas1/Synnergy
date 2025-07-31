// Synnergy Network – virtual_machine.go
// Build / run with:
//   go run ./cmd/virtual_machine --mode heavy --listen :9090
// -----------------------------------------------------------------------------
// NOTE: this file is self-contained; external packages (opcodes, common, logrus,
// wasmer-go) must still be in your go.mod.

package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/common" // common.Address
	"github.com/ethereum/go-ethereum/crypto" // crypto.Keccak256
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus" // aliased below
	"github.com/wasmerio/wasmer-go/wasmer"
	"golang.org/x/time/rate"
	"math/big"
	"net/http"
	"sync"
	"time"
)

const (
	PUSH Opcode = iota
	ADD
	STORE
	LOAD
	LOG
	RET

)

//---------------------------------------------------------------------
// Minimal state interface + in-memory implementation
//---------------------------------------------------------------------

type memState struct {
	mu         sync.RWMutex
	data       map[string][]byte
	balances   map[Address]uint64
	lpBalances map[Address]map[PoolID]uint64
	logs       []*Log
	contracts  map[Address][]byte
	tokens     map[TokenID]Token // <- Add this
	codeHashes map[Address]Hash
	nonces     map[Address]uint64
}

func (m *memState) Burn(addr Address, amt uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.balances[addr] < amt {
		return fmt.Errorf("insufficient balance: have %d, need %d", m.balances[addr], amt)
	}
	m.balances[addr] -= amt
	return nil
}

func (m *memState) BurnLP(addr Address, poolID PoolID, amt uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.lpBalances[addr]; !ok {
		return fmt.Errorf("no LP tokens found for address")
	}

	balance := m.lpBalances[addr][poolID]
	if balance < amt {
		return fmt.Errorf("insufficient LP balance")
	}

	m.lpBalances[addr][poolID] -= amt
	return nil
}

func (m *memState) MintLP(to Address, pool PoolID, amt uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.lpBalances[to]; !ok {
		m.lpBalances[to] = make(map[PoolID]uint64)
	}
	m.lpBalances[to][pool] += amt
	return nil
}

func NewInMemory() (StateRW, error) {
	return &memState{
		data:       make(map[string][]byte),
		balances:   make(map[Address]uint64),
		lpBalances: make(map[Address]map[PoolID]uint64),
		logs:       make([]*Log, 0),
		contracts:  make(map[Address][]byte),
		codeHashes: make(map[Address]Hash),
		nonces:     make(map[Address]uint64),
		tokens:     make(map[TokenID]Token),
	}, nil
}

func (m *memState) CallCode(from, to Address, input []byte, value *big.Int, gas uint64) ([]byte, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	code := m.contracts[to]
	if len(code) == 0 {
		return nil, false, fmt.Errorf("contract not found at %x", to)
	}

	// Wrap state to satisfy StateRW interface
	wrapper := &memStateWrapper{memState: m}

	ctx := &VMContext{
		Caller:   common.Address(from),
		TxHash:   sha256.Sum256(append(from[:], input...)),
		Code:     code,
		GasLimit: gas,
		State:    wrapper,
		Memory:   NewMemory(),
		GasMeter: NewGasMeter(gas),
	}

	// Select and execute appropriate VM
	var vm VM
	switch SelectVM(code) {
	case "superlight":
		vm = NewSuperLightVM(wrapper)
	case "light":
		vm = NewLightVM(wrapper, ctx.GasMeter)
	case "heavy":
		engine := wasmer.NewEngine()
		vm = NewHeavyVM(wrapper, ctx.GasMeter, engine)
	default:
		return nil, false, fmt.Errorf("unknown VM type selected")
	}

	receipt, err := vm.Execute(code, ctx)
	if err != nil {
		return nil, false, err
	}
	return receipt.ReturnData, true, nil
}

func (m *memState) CallContract(from, to Address, input []byte, value *big.Int, gas uint64) ([]byte, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	code := m.contracts[to]
	if len(code) == 0 {
		return nil, false, fmt.Errorf("contract not found at %x", to)
	}

	// Wrap for compatibility with StateRW interface
	wrapper := &memStateWrapper{memState: m}

	ctx := &VMContext{
		Caller:   common.Address(from),
		TxHash:   sha256.Sum256(append(from[:], input...)),
		Code:     code,
		GasLimit: gas,
		State:    wrapper,
		Memory:   NewMemory(),
		GasMeter: NewGasMeter(gas),
	}

	// Choose VM implementation
	var vm VM
	switch SelectVM(code) {
	case "superlight":
		vm = NewSuperLightVM(wrapper)
	case "light":
		vm = NewLightVM(wrapper, ctx.GasMeter)
	case "heavy":
		engine := wasmer.NewEngine()
		vm = NewHeavyVM(wrapper, ctx.GasMeter, engine)
	default:
		return nil, false, fmt.Errorf("unknown VM selected")
	}

	// Execute contract logic
	receipt, err := vm.Execute(code, ctx)
	if err != nil {
		return nil, false, err
	}
	return receipt.ReturnData, true, nil
}

func (m *memState) StaticCall(from, to Address, input []byte, gas uint64) ([]byte, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	code := m.contracts[to]
	if len(code) == 0 {
		return nil, false, fmt.Errorf("contract not found at %x", to)
	}

	wrapper := &memStateWrapper{memState: m}

	ctx := &VMContext{
		Caller:   common.Address(from),
		TxHash:   sha256.Sum256(append(from[:], input...)),
		Code:     code,
		GasLimit: gas,
		State:    wrapper,
		Memory:   NewMemory(),
		GasMeter: NewGasMeter(gas),
	}

	// Set static mode flag if needed in future
	// ctx.IsStatic = true

	var vm VM
	switch SelectVM(code) {
	case "superlight":
		vm = NewSuperLightVM(wrapper)
	case "light":
		vm = NewLightVM(wrapper, ctx.GasMeter)
	case "heavy":
		engine := wasmer.NewEngine()
		vm = NewHeavyVM(wrapper, ctx.GasMeter, engine)
	default:
		return nil, false, fmt.Errorf("unknown VM selected")
	}

	receipt, err := vm.Execute(code, ctx)
	if err != nil {
		return nil, false, err
	}
	return receipt.ReturnData, true, nil
}

func (m *memState) GetBalance(addr Address) uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.balances[addr]
}

func (m *memState) GetTokenBalance(addr Address, tokenID TokenID) (uint64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// For simplicity, return a dummy balance
	if m.balances == nil {
		return 0, fmt.Errorf("no balances initialized")
	}
	return m.balances[addr], nil
}

func (m *memState) SetTokenBalance(addr Address, tokenID TokenID, amount uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.balances == nil {
		m.balances = make(map[Address]uint64)
	}
	m.balances[addr] = amount
	return nil
}

func (m *memState) GetTokenSupply(tokenID TokenID) (uint64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// For simplicity, return a dummy supply
	if m.balances == nil {
		return 0, fmt.Errorf("no balances initialized")
	}
	var totalSupply uint64
	for _, balance := range m.balances {
		totalSupply += balance
	}
	return totalSupply, nil
}

func (m *memState) SetBalance(addr Address, amount uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.balances == nil {
		m.balances = make(map[Address]uint64)
	}
	m.balances[addr] = amount
	return nil
}

func (w *memStateWrapper) CallCode(from, to Address, input []byte, value *big.Int, gas uint64) ([]byte, bool, error) {
	return w.memState.CallCode(from, to, input, value, gas)
}

func (w *memStateWrapper) CallContract(from, to Address, input []byte, value *big.Int, gas uint64) ([]byte, bool, error) {
	return w.memState.CallContract(from, to, input, value, gas)
}

func (w *memStateWrapper) StaticCall(from, to Address, input []byte, gas uint64) ([]byte, bool, error) {
	return w.memState.StaticCall(from, to, input, gas)
}

func (m *memState) DelegateCall(from, to Address, input []byte, value *big.Int, gas uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	code := m.contracts[to]
	if len(code) == 0 {
		return fmt.Errorf("contract not found at %x", to)
	}

	// Wrap state to satisfy StateRW (if memState.Call returns [][]byte)
	wrapper := &memStateWrapper{memState: m}

	ctx := &VMContext{
		Caller:   common.Address(from), // ✅ fix for Address → common.Address
		TxHash:   sha256.Sum256(append(from[:], to[:]...)),
		Code:     code,
		GasLimit: gas,
		State:    wrapper, // ✅ use the wrapper
		Memory:   NewMemory(),
		GasMeter: NewGasMeter(gas),
	}

	// Select appropriate VM
	var vm VM
	switch SelectVM(code) {
	case "superlight":
		vm = NewSuperLightVM(wrapper)
	case "light":
		vm = NewLightVM(wrapper, ctx.GasMeter)
	case "heavy":
		engine := wasmer.NewEngine()
		vm = NewHeavyVM(wrapper, ctx.GasMeter, engine)
	default:
		return fmt.Errorf("unknown VM type selected")
	}

	_, err := vm.Execute(code, ctx)
	return err
}

func (m *memState) GetToken(tokenID TokenID) (Token, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	token, exists := m.tokens[tokenID]
	if !exists {
		return nil, fmt.Errorf("token with ID 0x%08X not found", tokenID)
	}
	return token, nil
}

func NewMemory() Memory {
	return &LinearMemory{
		data: make([]byte, 0, 1024),
	}
}

type LinearMemory struct {
	data []byte
}

func (m *LinearMemory) Read(offset, size uint64) []byte {
	end := offset + size
	if end > uint64(len(m.data)) {
		// Extend with zeroes
		newData := make([]byte, end)
		copy(newData, m.data)
		m.data = newData
	}
	return m.data[offset:end]
}

func (m *LinearMemory) Write(offset uint64, data []byte) {
	end := offset + uint64(len(data))
	if end > uint64(len(m.data)) {
		newData := make([]byte, end)
		copy(newData, m.data)
		m.data = newData
	}
	copy(m.data[offset:], data)
}

func (m *LinearMemory) Len() int {
	return len(m.data)
}

func (m *memState) Call(from, to Address, input []byte, value *big.Int, gas uint64) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	code := m.contracts[to]
	if len(code) == 0 {
		return nil, fmt.Errorf("contract not found at %x", to)
	}

	wrapper := &memStateWrapper{memState: m}

	ctx := &VMContext{
		Caller:   common.Address(from),
		TxHash:   sha256.Sum256(append(from[:], input...)),
		Code:     code,
		GasLimit: gas,
		State:    wrapper,
		Memory:   NewMemory(),
		GasMeter: NewGasMeter(gas),
	}

	vmType := SelectVM(code)
	var vm VM

	switch vmType {
	case "superlight":
		vm = NewSuperLightVM(wrapper)
	case "light":
		vm = NewLightVM(wrapper, ctx.GasMeter)
	case "heavy":
		engine := wasmer.NewEngine()
		vm = NewHeavyVM(wrapper, ctx.GasMeter, engine)
	default:
		return nil, fmt.Errorf("unknown VM type selected")
	}

	receipt, err := vm.Execute(code, ctx)
	if err != nil {
		return nil, fmt.Errorf("%s VM execution failed: %w", vmType, err)
	}

	return receipt.ReturnData, nil
}

type memStateWrapper struct {
	*memState
}

func (w *memStateWrapper) Call(from, to Address, input []byte, value *big.Int, gas uint64) ([]byte, error) {
	return w.memState.Call(from, to, input, value, gas)
}

func SelectVM(code []byte) string {
	if len(code) < 100 {
		return "superlight"
	} else if len(code) < 1000 {
		return "light"
	} else {
		return "heavy"
	}
}

func (m *memState) CreateContract(caller Address, code []byte, value *big.Int, gas uint64) (Address, []byte, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	nonce := m.nonces[caller]
	rlp := append(caller[:], byte(nonce))
	addrBytes := crypto.Keccak256(rlp)
	var contractAddr Address
	copy(contractAddr[:], addrBytes[:20])

	m.contracts[contractAddr] = code
	codeHash := sha256.Sum256(code)
	m.codeHashes[contractAddr] = codeHash
	m.nonces[caller]++

	commonCaller := common.BytesToAddress(caller[:])
	txHash := sha256.Sum256(append(caller[:], code...))

	wrapper := &memStateWrapper{memState: m}

	ctx := &VMContext{
		Caller:   commonCaller,
		TxHash:   txHash,
		Code:     code,
		GasLimit: gas,
		State:    wrapper,
		Memory:   NewMemory(),
		GasMeter: NewGasMeter(gas),
	}

	// 🔍 Select the appropriate VM
	vmType := SelectVM(code)
	var vm VM

	switch vmType {
	case "superlight":
		vm = NewSuperLightVM(wrapper)
	case "light":
		vm = NewLightVM(wrapper, ctx.GasMeter)
	case "heavy":
		engine := wasmer.NewEngine()
		vm = NewHeavyVM(wrapper, ctx.GasMeter, engine)
	default:
		return contractAddr, nil, false, fmt.Errorf("unknown VM type selected")
	}

	// ✅ Execute only the selected VM
	receipt, err := vm.Execute(code, ctx)
	if err != nil {
		return contractAddr, nil, false, fmt.Errorf("%s VM error: %w", vmType, err)
	}

	m.contracts[contractAddr] = receipt.ReturnData
	return contractAddr, receipt.ReturnData, true, nil
}

func (m *memState) GetContract(addr Address) (*Contract, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	codeKey := "code:" + addr.Hex()
	code, ok := m.data[codeKey]
	if !ok {
		return nil, fmt.Errorf("contract not found")
	}
	return &Contract{Address: addr, Bytecode: code}, nil
}

func (m *memState) AddLog(log *Log) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, log)
}

func (m *memState) GetCode(addr Address) []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.contracts[addr]
}

func (m *memState) GetCodeHash(addr Address) Hash {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.codeHashes[addr]
}

func (m *memState) MintToken(addr Address, amount uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.balances == nil {
		m.balances = make(map[Address]uint64)
	}
	m.balances[addr] += amount
	return nil
}

func (m *memState) Transfer(from, to Address, amount uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.balances == nil {
		return fmt.Errorf("no balances initialized")
	}
	if m.balances[from] < amount {
		return fmt.Errorf("insufficient balance: have %d, need %d", m.balances[from], amount)
	}
	m.balances[from] -= amount
	m.balances[to] += amount
	return nil
}

type memIterator struct {
	keys   [][]byte
	values [][]byte
	index  int
	err    error
}

func (m *memState) PrefixIterator(prefix []byte) StateIterator {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var keys [][]byte
	var values [][]byte

	for k, v := range m.data {
		if bytes.HasPrefix([]byte(k), prefix) {
			keys = append(keys, []byte(k))
			values = append(values, v)
		}
	}

	return &memIterator{
		keys:   keys,
		values: values,
		index:  -1,
	}
}

func (it *memIterator) Next() bool {
	it.index++
	return it.index < len(it.keys)
}

func (it *memIterator) Key() []byte {
	if it.index >= 0 && it.index < len(it.keys) {
		return it.keys[it.index]
	}
	return nil
}

func (it *memIterator) Value() []byte {
	if it.index >= 0 && it.index < len(it.values) {
		return it.values[it.index]
	}
	return nil
}

func (it *memIterator) Error() error {
	return it.err
}

func (m *memState) Snapshot(fn func() error) error {
	m.mu.Lock()
	origData := make(map[string][]byte, len(m.data))
	for k, v := range m.data {
		origData[k] = append([]byte(nil), v...)
	}
	origBalances := make(map[Address]uint64, len(m.balances))
	for k, v := range m.balances {
		origBalances[k] = v
	}
	origLP := make(map[Address]map[PoolID]uint64, len(m.lpBalances))
	for a, pools := range m.lpBalances {
		cp := make(map[PoolID]uint64, len(pools))
		for id, amt := range pools {
			cp[id] = amt
		}
		origLP[a] = cp
	}
	origContracts := make(map[Address][]byte, len(m.contracts))
	for a, c := range m.contracts {
		origContracts[a] = append([]byte(nil), c...)
	}
	origCodeHashes := make(map[Address]Hash, len(m.codeHashes))
	for a, h := range m.codeHashes {
		origCodeHashes[a] = h
	}
	origNonces := make(map[Address]uint64, len(m.nonces))
	for a, n := range m.nonces {
		origNonces[a] = n
	}
	origTokens := make(map[TokenID]Token, len(m.tokens))
	for id, t := range m.tokens {
		origTokens[id] = t
	}
	origLogs := append([]*Log(nil), m.logs...)
	err := fn()
	if err != nil {
		m.data = origData
		m.balances = origBalances
		m.lpBalances = origLP
		m.contracts = origContracts
		m.codeHashes = origCodeHashes
		m.nonces = origNonces
		m.tokens = origTokens
		m.logs = origLogs
	}
	m.mu.Unlock()
	return err
}

func (m *memState) NonceOf(addr Address) uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nonces[addr]
}

func (m *memState) IsIDTokenHolder(addr Address) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.balances[addr]
	return ok
}

func (m *memState) GetState(key []byte) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[string(key)]
	if !ok {
		return nil, fmt.Errorf("key not found")
	}
	return val, nil
}

func (m *memState) SetState(key, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[string(key)] = value
	return nil
}

func (m *memState) HasState(key []byte) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[string(key)]
	return ok, nil
}

func (m *memState) DeleteState(key []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := string(key)
	delete(m.data, k)
	return nil
}

func (m *memState) BalanceOf(addr Address) uint64 {
	if m.balances == nil {
		return 0
	}
	return m.balances[addr]
}

func (m *memState) Balance(addr Address) uint64 {
	if m.balances == nil {
		return 0
	}
	return m.balances[addr]
}
func (m *memState) Mint(addr Address, amt uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.balances == nil {
		m.balances = make(map[Address]uint64)
	}
	m.balances[addr] += amt
	return nil
}

func (m *memState) composite(ns, k []byte) string {
	return hex.EncodeToString(ns) + "|" + hex.EncodeToString(k)
}
func (m *memState) Get(ns, key []byte) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[m.composite(ns, key)]
	if !ok {
		return nil, errors.New("not found")
	}
	cpy := make([]byte, len(val))
	copy(cpy, val)
	return cpy, nil
}
func (m *memState) Set(ns, key, val []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cpy := make([]byte, len(val))
	copy(cpy, val)
	m.data[m.composite(ns, key)] = cpy
	return nil
}

//---------------------------------------------------------------------
// Context & Receipt types
//---------------------------------------------------------------------

type VMContext struct {
	Caller   common.Address
	Origin   common.Address
	TxHash   [32]byte
	GasLimit uint64
	Context
	Memory         Memory
	State          StateRW
	Chain          ChainContext
	PC             uint64
	JumpTable      map[uint64]struct{}
	GasMeter       *GasMeter
	LastReturnData []byte
	Code           []byte
}

// Memory is the linear byte‐array your opcodes read from and write to.
type Memory interface {
	// Read returns exactly `size` bytes (zero-extended if past the end).
	Read(offset, size uint64) []byte
	// Write writes the full slice at the given offset, growing as necessary.
	Write(offset uint64, data []byte)
	// Len returns the current size of memory.
	Len() int
}

// ChainContext provides the blockchain‐level data your opcodes need.
type ChainContext interface {
	BlockNumber() uint64
	Time() uint64
	Difficulty() *big.Int
	GasLimit() uint64
	ChainID() *big.Int
	BlockHash(number uint64) common.Hash
}

type Log struct {
	Address   Address       `json:"address"` // <- Add this
	Topics    []common.Hash `json:"topics"`  // <- Add this
	Data      []byte        `json:"data"`
	BlockTime int64         `json:"block_time"`
}

type Receipt struct {
	Status     bool   `json:"status"`
	GasUsed    uint64 `json:"gas_used"`
	ReturnData []byte `json:"return_data,omitempty"`
	Logs       []Log  `json:"logs,omitempty"`
	Error      string `json:"error,omitempty"`
}

//---------------------------------------------------------------------
// Helpers – gas, math, fail wrapper
//---------------------------------------------------------------------

// GasMeter tracks gas usage and enforces the execution gas limit.
type GasMeter struct {
	used  uint64 // gas consumed so far
	limit uint64 // total gas available
}

// NewGasMeter constructs a GasMeter with the given gas limit.
func NewGasMeter(limit uint64) *GasMeter {
	return &GasMeter{used: 0, limit: limit}
}

func (m *memState) SelfDestruct(contract, beneficiary Address) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Transfer contract balance to beneficiary
	balance := m.balances[contract]
	m.balances[contract] = 0
	m.balances[beneficiary] += balance

	// Remove contract code and code hash
	delete(m.contracts, contract)
	delete(m.codeHashes, contract)

	// Optionally delete state or tokens tied to the contract
}

// Remaining returns the current gas remaining.
func (g *GasMeter) Remaining() uint64 {
	return g.limit - g.used
}

func (g *GasMeter) Consume(op Opcode) error {
	c := GasCost(op)
	if g.used+c > g.limit {
		return fmt.Errorf("out-of-gas (%d/%d)", g.used+c, g.limit)
	}
	g.used += c
	return nil
}

// AddBigInts – deterministic addition for arbitrary-length byte slices.
func AddBigInts(a, b []byte) []byte {
	var ai, bi big.Int
	ai.SetBytes(a)
	bi.SetBytes(b)
	return new(big.Int).Add(&ai, &bi).Bytes()
}

//---------------------------------------------------------------------
// VM interface + three implementations
//---------------------------------------------------------------------

type VM interface {
	Execute(bytecode []byte, ctx *VMContext) (*Receipt, error)
}

type SuperLightVM struct{ led StateRW }
type LightVM struct {
	led StateRW
	gas *GasMeter
}
type HeavyVM struct {
	led    StateRW
	gas    *GasMeter
	engine *wasmer.Engine
}

func NewSuperLightVM(led StateRW) VM {
	return &SuperLightVM{led: led}
}

func NewLightVM(led StateRW, gas *GasMeter) VM {
	return &LightVM{led: led, gas: gas}
}

func NewHeavyVM(led StateRW, gas *GasMeter, engine *wasmer.Engine) VM {
	return &HeavyVM{led: led, gas: gas, engine: engine}
}

//---------------------------------------------------------------------
// Super-Light (sig / nonce check only)
//---------------------------------------------------------------------

func (vm *SuperLightVM) Execute(bc []byte, ctx *VMContext) (*Receipt, error) {
	if sha256.Sum256(bc) != ctx.TxHash {
		return &Receipt{Status: false, Error: "tx hash mismatch"}, nil
	}
	return &Receipt{Status: true, GasUsed: 0}, nil
}

//---------------------------------------------------------------------
// Light interpreter
//---------------------------------------------------------------------

func (vm *LightVM) Execute(b []byte, ctx *VMContext) (*Receipt, error) {
	rec := &Receipt{Status: true}
	stack := make([][]byte, 0, 16)
	pc := 0
	meter := vm.gas
	store := vm.led

	push := func(d []byte) { stack = append(stack, d) }
	pop := func() ([]byte, error) {
		if len(stack) == 0 {
			return nil, errors.New("stack underflow")
		}
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return v, nil
	}

	for pc < len(b) {
		op := Opcode(b[pc])
		pc++
		if err := meter.Consume(op); err != nil {
			return fail(rec, err)
		}

		switch op {
		case PUSH:
			if pc >= len(b) {
				return fail(rec, errors.New("missing length byte"))
			}
			l := int(b[pc])
			pc++
			if pc+l > len(b) {
				return fail(rec, errors.New("push out of bounds"))
			}
			push(b[pc : pc+l])
			pc += l

		case ADD:
			a, err := pop()
			if err != nil {
				return fail(rec, err)
			}
			b1, err := pop()
			if err != nil {
				return fail(rec, err)
			}
			push(AddBigInts(a, b1))

		case STORE:
			key, err := pop()
			if err != nil {
				return fail(rec, err)
			}
			val, err := pop()
			if err != nil {
				return fail(rec, err)
			}
			if err := store.Set(ctx.TxHash[:], key, val); err != nil {
				return fail(rec, err)
			}

		case LOAD:
			key, err := pop()
			if err != nil {
				return fail(rec, err)
			}
			val, err := store.Get(ctx.TxHash[:], key)
			if err != nil {
				return fail(rec, err)
			}
			push(val)

		case LOG:
			msg, err := pop()
			if err != nil {
				return fail(rec, err)
			}
			rec.Logs = append(rec.Logs, Log{
				BlockTime: time.Now().Unix(),
				Topics:    []common.Hash{common.BytesToHash(ctx.TxHash[:])},
				Data:      msg,
			})

		case RET:
			rd, _ := pop()
			rec.ReturnData = rd
			rec.GasUsed = meter.used
			return rec, nil

		default:
			return fail(rec, fmt.Errorf("unknown opcode 0x%02X", op))
		}
	}
	rec.GasUsed = meter.used
	return rec, nil
}

//---------------------------------------------------------------------
// Heavy (Wasmer JIT) – host bindings
//---------------------------------------------------------------------

type hostCtx struct {
	mem   *wasmer.Memory
	store StateRW
	gas   *GasMeter
	tx    *VMContext
	rec   *Receipt
}

func (vm *HeavyVM) Execute(code []byte, ctx *VMContext) (*Receipt, error) {
	rec := &Receipt{Status: true}

	store := wasmer.NewStore(vm.engine) // ← you already have this
	mod, err := wasmer.NewModule(store, code)
	if err != nil {
		return nil, err
	}

	hctx := &hostCtx{store: vm.led, gas: vm.gas, tx: ctx, rec: rec}

	imports := registerHost(store, hctx) // ← pass store **and** hctx

	instance, err := wasmer.NewInstance(mod, imports)
	if err != nil {
		return nil, err
	}

	mem, err := instance.Exports.GetMemory("memory")
	if err != nil {
		return nil, errors.New("wasm memory export missing")
	}
	hctx.mem = mem

	start, err := instance.Exports.GetFunction("_start")
	if err != nil {
		return nil, errors.New("_start function required")
	}
	if _, err = start(); err != nil {
		rec.Status = false
		rec.Error = err.Error()
	}

	rec.GasUsed = vm.gas.used
	return rec, nil
}

// registerHost converts your Go callbacks into Wasm imports.
// `store` is the same store used to compile the module.
func registerHost(store *wasmer.Store, h *hostCtx) *wasmer.ImportObject {
	imports := wasmer.NewImportObject()

	// -----------------------------------------------------------------
	// re-usable helpers for memory access
	// -----------------------------------------------------------------
	read := func(ptr, ln int32) []byte {
		bytes := h.mem.Data()[ptr : ptr+ln]
		out := make([]byte, ln)
		copy(out, bytes)
		return out
	}
	write := func(ptr int32, data []byte) { copy(h.mem.Data()[ptr:], data) }

	// -----------------------------------------------------------------
	// host_consume_gas(op u32) -> i32
	// -----------------------------------------------------------------
	hostConsumeGas := wasmer.NewFunction(
		store,
		wasmer.NewFunctionType(
			// cast the C-constant into a Go ValueKind:
			wasmer.NewValueTypes(wasmer.ValueKind(wasmer.I32)),
			wasmer.NewValueTypes(wasmer.ValueKind(wasmer.I32)),
		),
		func(args []wasmer.Value) ([]wasmer.Value, error) {
			op := uint32(args[0].I32())
			if err := h.gas.Consume(Opcode(op)); err != nil {
				h.rec.Status = false
				h.rec.Error = err.Error()
				return []wasmer.Value{wasmer.NewI32(-1)}, nil
			}
			return []wasmer.Value{wasmer.NewI32(0)}, nil
		},
	)

	// -----------------------------------------------------------------
	// host_read(keyPtr,len,dstPtr) -> i32(len)|-1
	// -----------------------------------------------------------------
	hostRead := wasmer.NewFunction(
		store,
		wasmer.NewFunctionType(
			// cast each I32 into a ValueKind
			wasmer.NewValueTypes(
				wasmer.ValueKind(wasmer.I32),
				wasmer.ValueKind(wasmer.I32),
				wasmer.ValueKind(wasmer.I32),
			),
			wasmer.NewValueTypes(
				wasmer.ValueKind(wasmer.I32),
			),
		),
		func(args []wasmer.Value) ([]wasmer.Value, error) {
			kPtr, kLen, dPtr := args[0].I32(), args[1].I32(), args[2].I32()
			key := read(kPtr, kLen)
			val, err := h.store.Get(h.tx.TxHash[:], key)
			if err != nil {
				return []wasmer.Value{wasmer.NewI32(-1)}, nil
			}
			write(dPtr, val)
			return []wasmer.Value{wasmer.NewI32(int32(len(val)))}, nil
		},
	)

	// -----------------------------------------------------------------
	// host_write(keyPtr,len,valPtr,valLen) -> i32
	// -----------------------------------------------------------------
	hostWrite := wasmer.NewFunction(
		store,
		wasmer.NewFunctionType(
			// four i32 params, one i32 result
			wasmer.NewValueTypes(
				wasmer.ValueKind(wasmer.I32),
				wasmer.ValueKind(wasmer.I32),
				wasmer.ValueKind(wasmer.I32),
				wasmer.ValueKind(wasmer.I32),
			),
			wasmer.NewValueTypes(
				wasmer.ValueKind(wasmer.I32),
			),
		),
		func(args []wasmer.Value) ([]wasmer.Value, error) {
			kPtr, kLen, vPtr, vLen := args[0].I32(), args[1].I32(), args[2].I32(), args[3].I32()
			key := read(kPtr, kLen)
			val := read(vPtr, vLen)
			if err := h.store.Set(h.tx.TxHash[:], key, val); err != nil {
				h.rec.Status = false
				h.rec.Error = err.Error()
				return []wasmer.Value{wasmer.NewI32(-1)}, nil
			}
			return []wasmer.Value{wasmer.NewI32(0)}, nil
		},
	)

	// -----------------------------------------------------------------
	// host_log(ptr,len)
	// -----------------------------------------------------------------
	hostLog := wasmer.NewFunction(
		store,
		wasmer.NewFunctionType(
			// two i32 params
			wasmer.NewValueTypes(
				wasmer.ValueKind(wasmer.I32),
				wasmer.ValueKind(wasmer.I32),
			),
			// no results
			wasmer.NewValueTypes(),
		),
		func(args []wasmer.Value) ([]wasmer.Value, error) {
			p, l := args[0].I32(), args[1].I32()
			msg := read(p, l)
			h.rec.Logs = append(h.rec.Logs, Log{
				BlockTime: time.Now().Unix(),
				Topics:    []common.Hash{common.BytesToHash(h.tx.TxHash[:])},
				Data:      msg,
			})
			// return an empty slice for no results:
			return []wasmer.Value{}, nil
		},
	)

	// Register all functions under the "env" namespace.
	imports.Register("env", map[string]wasmer.IntoExtern{
		"host_consume_gas": hostConsumeGas,
		"host_read":        hostRead,
		"host_write":       hostWrite,
		"host_log":         hostLog,
	})

	return imports
}

//---------------------------------------------------------------------
// HTTP API + rate limiter
//---------------------------------------------------------------------

var limiter = rate.NewLimiter(200, 100) // 200 req/s, burst 100

func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "rate limit", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

//---------------------------------------------------------------------
// Bootstrap
//---------------------------------------------------------------------

func main() {
	mode := flag.String("mode", "super-light", "super-light | light | heavy")
	listen := flag.String("listen", ":9090", "listen address")
	gas := flag.Uint64("gas", 8_000_000, "default gas limit")
	flag.Parse()

	logrus.SetFormatter(&logrus.JSONFormatter{})

	led, _ := NewInMemory()

	var engine VM
	switch *mode {
	case "super-light":
		engine = &SuperLightVM{led}
	case "light":
		engine = &LightVM{led, NewGasMeter(*gas)}
	case "heavy":
		engine = &HeavyVM{led, NewGasMeter(*gas), wasmer.NewEngine()}
	default:
		logrus.Fatal("invalid mode")
	}

	r := mux.NewRouter()
	r.Use(limit)
	r.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Code string    `json:"bytecode"`
			Ctx  VMContext `json:"ctx"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		code, err := hex.DecodeString(req.Code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		rec, err := engine.Execute(code, &req.Ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rec)
	}).Methods("POST")

	srv := &http.Server{
		Addr:         *listen,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	logrus.Infof("VM %s listening on %s", *mode, *listen)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logrus.Fatal(err)
	}
}

func fail(rec *Receipt, err error) (*Receipt, error) {
	rec.Status = false
	rec.Error = err.Error()
	return rec, err
}
