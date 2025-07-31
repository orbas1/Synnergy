// Code generated for Synnergy Network – tokens.go
// Author: ChatGPT (OpenAI o3)
// Description: Core token registry, universal token factory, and VM opcode integration
//
//	Instantiates the 50 canonical assets defined in the Synthron Token Guide.
//
// -----------------------------------------------------------------------------
package core

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sort"
	"sync"
	"time"
)

//---------------------------------------------------------------------
// Token‑ID & Standard byte map
//---------------------------------------------------------------------

type TokenID uint32

var ErrInvalidAsset = errors.New("invalid token asset")

const (
	OpTokenTransfer = 0xB0
	OpTokenApprove  = 0xB1
	OpAllowance     = 0xB2
	OpBalanceOf     = 0xB3
)

var AddressZero Address

const (
	StdSYN20   byte = 0x14
	StdSYN70   byte = 0x46
	StdSYN130  byte = 0x82
	StdSYN131  byte = 0x83
	StdSYN200  byte = 0x32
	StdSYN223  byte = 0xDF
	StdSYN300  byte = 0x4B
	StdSYN500  byte = 0x4D
	StdSYN600  byte = 0x4E
	StdSYN700  byte = 0x57
	StdSYN721  byte = 0xD1
	StdSYN722  byte = 0xD2
	StdSYN800  byte = 0x50
	StdSYN845  byte = 0xED
	StdSYN900  byte = 0x52
	StdSYN1000 byte = 0x58
	StdSYN1100 byte = 0x56
	StdSYN1155 byte = 0x92
	StdSYN1200 byte = 0x5A
	StdSYN1300 byte = 0x66
	StdSYN1401 byte = 0xF5
	StdSYN1500 byte = 0x5F
	StdSYN1600 byte = 0x68
	StdSYN1700 byte = 0x6A
	StdSYN1800 byte = 0x6C
	StdSYN1900 byte = 0x6E
	StdSYN1967 byte = 0x66
	StdSYN2100 byte = 0x70
	StdSYN2200 byte = 0x71
	StdSYN2369 byte = 0x9A
	StdSYN2400 byte = 0x72
	StdSYN2500 byte = 0x73
	StdSYN2600 byte = 0x74
	StdSYN2700 byte = 0x75
	StdSYN2800 byte = 0x76
	StdSYN2900 byte = 0x77
	StdSYN3000 byte = 0x78
	StdSYN3100 byte = 0x79
	StdSYN3200 byte = 0x7A
	StdSYN3300 byte = 0x7B
	StdSYN3400 byte = 0x7C
	StdSYN3500 byte = 0x7D
	StdSYN3600 byte = 0x7E
	StdSYN3700 byte = 0x7F
	StdSYN3800 byte = 0x80
	StdSYN3900 byte = 0x81
	StdSYN4200 byte = 0x84
	StdSYN4300 byte = 0x85
	StdSYN4700 byte = 0x86
	StdSYN4900 byte = 0x87
	StdSYN5000 byte = 0x88
	StdSYN10   byte = 0x0A
)

//---------------------------------------------------------------------
// Metadata & Token interface
//---------------------------------------------------------------------

type Metadata struct {
	Name        string
	Symbol      string
	Decimals    uint8
	Standard    byte
	Created     time.Time
	FixedSupply bool
	TotalSupply uint64
}

type Token interface {
	ID() TokenID
	Meta() Metadata
	BalanceOf(addr Address) uint64
	Transfer(from, to Address, amount uint64) error
	Allowance(owner, spender Address) uint64
	Approve(owner, spender Address, amount uint64) error
	Mint(to Address, amount uint64) error
	Burn(from Address, amount uint64) error
}

func (t *BaseToken) Mint(to Address, amount uint64) error {
	if t.balances == nil {
		t.balances = &BalanceTable{}
	}
	t.balances.Add(t.id, to, amount)
	return nil
}

func (t *BaseToken) Burn(from Address, amount uint64) error {
	if t.balances == nil {
		return fmt.Errorf("balances not initialized")
	}
	return t.balances.Sub(t.id, from, amount)
}

func (bt *BalanceTable) Add(tokenID TokenID, to Address, amount uint64) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if bt.balances == nil {
		bt.balances = make(map[TokenID]map[Address]uint64)
	}

	if bt.balances[tokenID] == nil {
		bt.balances[tokenID] = make(map[Address]uint64)
	}

	bt.balances[tokenID][to] += amount
}

func (bt *BalanceTable) Sub(tokenID TokenID, from Address, amount uint64) error {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if bt.balances == nil || bt.balances[tokenID] == nil {
		return fmt.Errorf("no balance to subtract from")
	}

	if bt.balances[tokenID][from] < amount {
		return fmt.Errorf("insufficient balance")
	}

	bt.balances[tokenID][from] -= amount
	return nil
}

//---------------------------------------------------------------------
// BaseToken implementation (unchanged)
//---------------------------------------------------------------------

type BaseToken struct {
	id        TokenID
	meta      Metadata
	balances  *BalanceTable
	allowance sync.Map
	lock      sync.RWMutex
	ledger    *Ledger
	gas       GasCalculator
}

func (b *BaseToken) ID() TokenID    { return b.id }
func (b *BaseToken) Meta() Metadata { return b.meta }

func (b *BaseToken) BalanceOf(a Address) uint64 { return b.balances.Get(b.id, a) }

func (b *BaseToken) Allowance(o, s Address) uint64 {
	if v, ok := b.allowance.Load(o); ok {
		if inner, ok2 := v.(*sync.Map); ok2 {
			if amt, ok3 := inner.Load(s); ok3 {
				return amt.(uint64)
			}
		}
	}
	return 0
}

func (bt *BalanceTable) Get(tokenID TokenID, addr Address) uint64 {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	if bt.balances == nil {
		return 0
	}

	tokenBalances, ok := bt.balances[tokenID]
	if !ok {
		return 0
	}

	return tokenBalances[addr]
}

func (b *BaseToken) Approve(o, s Address, amt uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	inner, _ := b.allowance.LoadOrStore(o, &sync.Map{})
	inner.(*sync.Map).Store(s, amt)
	b.ledger.EmitApproval(b.id, o, s, amt)
	log.WithFields(log.Fields{"token": b.meta.Symbol, "owner": o, "spender": s, "amount": amt}).Info("approve")
	return nil
}
func (b *BaseToken) Transfer(from, to Address, amt uint64) error {
	if err := b.ledger.WithinBlock(func() error {
		if err := b.balances.Sub(b.id, from, amt); err != nil {
			return err
		}
		b.balances.Add(b.id, to, amt)
		return nil
	}); err != nil {
		return err
	}
	fee := b.gas.Calculate("OpTokenTransfer", amt)
	b.ledger.DeductGas(from, fee)
	b.ledger.EmitTransfer(b.id, from, to, amt)
	log.WithFields(log.Fields{"token": b.meta.Symbol, "from": from, "to": to, "amount": amt, "gas": fee}).Info("transfer")
	return nil
}

func (Calculator) Calculate(op byte, amount uint64) uint64 {
	switch op {
	case OpTokenTransfer:
		return 500 + amount/10000
	default:
		return 0
	}
}

type Calculator struct{}

//---------------------------------------------------------------------
// Registry singleton
//---------------------------------------------------------------------

var (
	regOnce sync.Once
)

func getRegistry() *ContractRegistry {
	regOnce.Do(func() {
		if reg == nil {
			reg = &ContractRegistry{
				Registry: &Registry{
					Entries: make(map[string][]byte),
					tokens:  make(map[TokenID]*BaseToken),
				},
				byAddr: make(map[Address]*SmartContract),
			}
		} else {
			if reg.Registry == nil {
				reg.Registry = &Registry{Entries: make(map[string][]byte), tokens: make(map[TokenID]*BaseToken)}
			}
			if reg.byAddr == nil {
				reg.byAddr = make(map[Address]*SmartContract)
			}
			if reg.Registry.tokens == nil {
				reg.Registry.tokens = make(map[TokenID]*BaseToken)
			}
		}
	})
	return reg
}

func RegisterToken(t Token) {
	r := getRegistry()
	r.mu.Lock()
	if r.Registry == nil {
		r.Registry = &Registry{Entries: make(map[string][]byte), tokens: make(map[TokenID]*BaseToken)}
	}
	if r.Registry.tokens == nil {
		r.Registry.tokens = make(map[TokenID]*BaseToken)
	}
	r.Registry.tokens[t.ID()] = t.(*BaseToken)
	r.mu.Unlock()
	log.WithField("symbol", t.Meta().Symbol).Info("token registered")
}

func GetToken(id TokenID) (Token, bool) {
	r := getRegistry()
	r.mu.RLock()
	defer r.mu.RUnlock()
	tok, ok := r.Registry.tokens[id]
	return tok, ok
}

func GetRegistryTokens() []*BaseToken {
	r := getRegistry()
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*BaseToken, 0, len(r.Registry.tokens))
	for _, t := range r.Registry.tokens {
		list = append(list, t)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].id < list[j].id })
	return list
}

func InitTokens(ledger *Ledger, vm VM, gas GasCalculator) {
	r := getRegistry()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.vm = vm
	if ledger.tokens == nil {
		ledger.tokens = make(map[TokenID]Token)
	}
	for id, tok := range r.Registry.tokens {
		tok.ledger = ledger
		tok.gas = gas
		ledger.tokens[id] = tok
	}
}

//---------------------------------------------------------------------
// Factory
//---------------------------------------------------------------------

type Factory struct{}

func (Factory) Create(meta Metadata, init map[Address]uint64) (Token, error) {
	if meta.Created.IsZero() {
		meta.Created = time.Now().UTC()
	}
	bt := &BaseToken{id: deriveID(meta.Standard), meta: meta, balances: NewBalanceTable()}
	for a, v := range init {
		bt.balances.Set(bt.id, a, v)
		bt.meta.TotalSupply += v
	}
	RegisterToken(bt)
	return bt, nil
}

func NewBalanceTable() *BalanceTable {
	return &BalanceTable{
		balances: make(map[TokenID]map[Address]uint64),
	}
}

func (bt *BalanceTable) Set(tokenID TokenID, addr Address, amount uint64) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if bt.balances == nil {
		bt.balances = make(map[TokenID]map[Address]uint64)
	}

	if bt.balances[tokenID] == nil {
		bt.balances[tokenID] = make(map[Address]uint64)
	}

	bt.balances[tokenID][addr] = amount
}

func deriveID(std byte) TokenID { return TokenID(0x53000000 | uint32(std)<<8) }

//---------------------------------------------------------------------
// Canonical token mint (50 assets)
//---------------------------------------------------------------------

func init() {
	f := Factory{}
	canon := []Metadata{
		{"Synnergy Native", "SYN", 18, StdSYN20, time.Time{}, false, 0},
		{"Synnergy Governance", "SYN-GOV", 18, StdSYN300, time.Time{}, true, 0},
		{"Synnergy Stable USD", "SYNUSD", 6, StdSYN1000, time.Time{}, false, 0},
		{"Synnergy Carbon Credit", "SYN-CO2", 0, StdSYN200, time.Time{}, false, 0},
		{"Synnergy SupplyChain", "SYNSC", 0, StdSYN1300, time.Time{}, false, 0},
		{"Synnergy Music Royalty", "SYN-MUSIC", 0, StdSYN1600, time.Time{}, false, 0},
		{"Synnergy Healthcare", "SYN-HDATA", 0, StdSYN1100, time.Time{}, false, 0},
		{"Synnergy IP", "SYN-IP", 0, StdSYN700, time.Time{}, false, 0},
		{"Synnergy Gold", "SYN-GOLD", 3, StdSYN1967, time.Time{}, false, 0},
		{"Synnergy Oil", "SYN-OIL", 2, StdSYN1967, time.Time{}, false, 0},
		{"Synnergy Reputation", "SYN-REP", 0, StdSYN1500, time.Time{}, false, 0},
		{"Synnergy Interop", "SYNX", 18, StdSYN1200, time.Time{}, false, 0},
		{"Synnergy NFT Art", "SYNART", 0, StdSYN721, time.Time{}, false, 0},
		{"Synnergy NFT Land", "SYNLAND", 0, StdSYN2369, time.Time{}, false, 0},
		{"Synnergy Ticket", "SYNTIX", 0, StdSYN1700, time.Time{}, false, 0},
		{"Synnergy Debt", "SYN-LOAN", 0, StdSYN845, time.Time{}, false, 0},
		{"Synnergy Reward", "SYN-RWD", 18, StdSYN600, time.Time{}, false, 0},
		{"Synnergy Utility", "SYN-UTIL", 18, StdSYN500, time.Time{}, false, 0},
		{"Synnergy Game", "SYNGAME", 0, StdSYN70, time.Time{}, false, 0},
		{"Synnergy Multi-Asset", "SYN-MA", 0, StdSYN1155, time.Time{}, false, 0},
		{"Synnergy Bond", "SYN-BOND", 0, StdSYN1401, time.Time{}, false, 0},
		{"Synnergy Tangible", "SYN-TANG", 0, StdSYN130, time.Time{}, false, 0},
		{"Synnergy Intangible", "SYN-INTANG", 0, StdSYN131, time.Time{}, false, 0},
		{"Synnergy SafeTransfer", "SYN223", 18, StdSYN223, time.Time{}, false, 0},
		{"Synnergy Identity", "SYN-ID", 0, StdSYN900, time.Time{}, false, 0},
		{"Synnergy CBDC", "SYN-CBDC", 2, StdSYN10, time.Time{}, false, 0},
		{"Synnergy Asset‑Backed", "SYN-ASSET", 0, StdSYN800, time.Time{}, false, 0},
		{"Synnergy ETF", "SYN-ETF", 0, StdSYN3300, time.Time{}, false, 0},
		{"Synnergy Forex", "SYN-FX", 0, StdSYN3400, time.Time{}, false, 0},
		{"Synnergy Currency", "SYN-CUR", 0, StdSYN3500, time.Time{}, false, 0},
		{"Synnergy Futures", "SYN-FUT", 0, StdSYN3600, time.Time{}, false, 0},
		{"Synnergy Index", "SYN-INDEX", 0, StdSYN3700, time.Time{}, false, 0},
		{"Synnergy Grant", "SYN-GRANT", 0, StdSYN3800, time.Time{}, false, 0},
		{"Synnergy Benefit", "SYN-BEN", 0, StdSYN3900, time.Time{}, false, 0},
		{"Synnergy Charity", "SYN-CHRTY", 0, StdSYN4200, time.Time{}, false, 0},
		{"Synnergy Energy", "SYN-ENRG", 0, StdSYN4300, time.Time{}, false, 0},
		{"Synnergy Legal", "SYN-LEGAL", 0, StdSYN4700, time.Time{}, false, 0},
		{"Synnergy Agriculture", "SYN-AGRI", 0, StdSYN4900, time.Time{}, false, 0},
		{"Synnergy Carbon Footprint", "SYN-CFP", 0, StdSYN1800, time.Time{}, false, 0},
		{"Synnergy Education", "SYN-EDU", 0, StdSYN1900, time.Time{}, false, 0},
		{"Synnergy Supply‑Fin", "SYN-SCFIN", 0, StdSYN2100, time.Time{}, false, 0},
		{"Synnergy RTP", "SYN-RTP", 0, StdSYN2200, time.Time{}, false, 0},
		{"Synnergy Data", "SYN-DATA", 0, StdSYN2400, time.Time{}, false, 0},
		{"Synnergy DAO", "SYN-DAO", 0, StdSYN2500, time.Time{}, false, 0},
		{"Synnergy Investor", "SYN-INV", 0, StdSYN2600, time.Time{}, false, 0},
		{"Synnergy Pension", "SYN-PENS", 0, StdSYN2700, time.Time{}, false, 0},
		{"Synnergy Life", "SYN-LIFE", 0, StdSYN2800, time.Time{}, false, 0},
		{"Synnergy Insurance", "SYN-INSUR", 0, StdSYN2900, time.Time{}, false, 0},
		{"Synnergy Rental", "SYN-RENT", 0, StdSYN3000, time.Time{}, false, 0},
		{"Synnergy Employment", "SYN-EMP", 0, StdSYN3100, time.Time{}, false, 0},
		{"Synnergy Bill", "SYN-BILL", 0, StdSYN3200, time.Time{}, false, 0},
	}

	for _, m := range canon {
		if _, err := f.Create(m, map[Address]uint64{AddressZero: 0}); err != nil {
			panic(err)
		}
	}

	registerTokenOpcodes()
}

//---------------------------------------------------------------------
// VM opcode binding – basic transfer shown, others omitted for brevity
//---------------------------------------------------------------------

func registerTokenOpcodes() {
	Register(0xB0, func(ctx OpContext) error {
		// Delegate actual logic to the VM environment if available.
		return ctx.Call("Tokens_Transfer")
	})
	// APPROVE 0xB1, ALLOWANCE 0xB2, BALANCEOF 0xB3 can be registered similarly.
}

func (ctx *Context) RefundGas(amount uint64) {
	ctx.GasPrice += amount
}

type Stack struct {
	data []any
}

func (s *Stack) PopUint32() uint32 {
	if len(s.data) == 0 {
		panic("stack underflow")
	}

	v := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]

	switch val := v.(type) {
	case uint32:
		return val
	case uint64:
		return uint32(val)
	case int:
		return uint32(val)
	default:
		panic("invalid type for PopUint32")
	}
}

func (s *Stack) PopAddress() Address {
	if len(s.data) == 0 {
		panic("stack underflow: PopAddress")
	}

	v := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]

	addr, ok := v.(Address)
	if !ok {
		panic("invalid type on stack: expected Address")
	}

	return addr
}

func (s *Stack) PopUint64() uint64 {
	if len(s.data) == 0 {
		panic("stack underflow: PopUint64")
	}

	v := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]

	switch val := v.(type) {
	case uint64:
		return val
	case int:
		return uint64(val)
	case uint32:
		return uint64(val)
	default:
		panic("invalid type on stack: expected uint64")
	}
}

func (s *Stack) PushBool(b bool) {
	s.data = append(s.data, b)
}

func (s *Stack) Push(v any) {
	s.data = append(s.data, v)
}

func (s *Stack) Len() int {
	return len(s.data)
}
