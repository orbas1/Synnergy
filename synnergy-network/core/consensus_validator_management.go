package core

import (
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// ValidatorInfo represents a consensus validator and its staked amount.
type ValidatorInfo struct {
	Addr     Address `json:"addr"`
	Stake    uint64  `json:"stake"`
	Active   bool    `json:"active"`
	JoinedAt int64   `json:"since"`
}

// ValidatorManager keeps track of validators and their stakes.
type ValidatorManager struct {
	mu     sync.RWMutex
	ledger StateRW
}

var (
	// StakingAccount holds locked validator stakes.
	StakingAccount Address
)

func init() {
	var err error
	StakingAccount, err = StringToAddress("0x5374616b696e674163636f756e74000000000000")
	if err != nil {
		panic("invalid StakingAccount: " + err.Error())
	}
}

// NewValidatorManager constructs a manager with the provided ledger backend.
func NewValidatorManager(led StateRW) *ValidatorManager { return &ValidatorManager{ledger: led} }

// Register adds a validator and locks the initial stake.
func (vm *ValidatorManager) Register(addr Address, stake uint64) error {
	if stake == 0 {
		return errors.New("stake must be >0")
	}
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if ok, _ := vm.ledger.HasState(vm.key(addr)); ok {
		return errors.New("already registered")
	}
	if err := vm.ledger.Transfer(addr, StakingAccount, stake); err != nil {
		return err
	}
	info := ValidatorInfo{Addr: addr, Stake: stake, Active: true, JoinedAt: time.Now().Unix()}
	b, _ := json.Marshal(info)
	vm.ledger.SetState(vm.key(addr), b)
	return nil
}

// Deregister removes a validator and returns its stake.
func (vm *ValidatorManager) Deregister(addr Address) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	raw, err := vm.ledger.GetState(vm.key(addr))
	if err != nil || len(raw) == 0 {
		return errors.New("not registered")
	}
	var info ValidatorInfo
	_ = json.Unmarshal(raw, &info)
	if err := vm.ledger.Transfer(StakingAccount, addr, info.Stake); err != nil {
		return err
	}
	info.Active = false
	info.Stake = 0
	vm.ledger.DeleteState(vm.key(addr))
	return nil
}

// Stake increases a validator's locked stake.
func (vm *ValidatorManager) Stake(addr Address, amt uint64) error {
	if amt == 0 {
		return errors.New("amount must be >0")
	}
	vm.mu.Lock()
	defer vm.mu.Unlock()
	raw, err := vm.ledger.GetState(vm.key(addr))
	if err != nil || len(raw) == 0 {
		return errors.New("not registered")
	}
	var info ValidatorInfo
	_ = json.Unmarshal(raw, &info)
	if err := vm.ledger.Transfer(addr, StakingAccount, amt); err != nil {
		return err
	}
	info.Stake += amt
	b, _ := json.Marshal(info)
	vm.ledger.SetState(vm.key(addr), b)
	return nil
}

// Unstake releases a portion of a validator's stake back to the owner.
func (vm *ValidatorManager) Unstake(addr Address, amt uint64) error {
	if amt == 0 {
		return errors.New("amount must be >0")
	}
	vm.mu.Lock()
	defer vm.mu.Unlock()
	raw, err := vm.ledger.GetState(vm.key(addr))
	if err != nil || len(raw) == 0 {
		return errors.New("not registered")
	}
	var info ValidatorInfo
	_ = json.Unmarshal(raw, &info)
	if info.Stake < amt {
		return errors.New("insufficient stake")
	}
	if err := vm.ledger.Transfer(StakingAccount, addr, amt); err != nil {
		return err
	}
	info.Stake -= amt
	b, _ := json.Marshal(info)
	vm.ledger.SetState(vm.key(addr), b)
	return nil
}

// Slash deducts stake as a penalty. Burned amounts reduce total supply.
func (vm *ValidatorManager) Slash(addr Address, amt uint64) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	raw, err := vm.ledger.GetState(vm.key(addr))
	if err != nil || len(raw) == 0 {
		return errors.New("not registered")
	}
	var info ValidatorInfo
	_ = json.Unmarshal(raw, &info)
	if amt > info.Stake {
		amt = info.Stake
	}
	if amt > 0 {
		if err := vm.ledger.Burn(StakingAccount, amt); err != nil {
			return err
		}
		info.Stake -= amt
	}
	if info.Stake == 0 {
		info.Active = false
	}
	b, _ := json.Marshal(info)
	vm.ledger.SetState(vm.key(addr), b)
	return nil
}

// Get returns information for a validator.
func (vm *ValidatorManager) Get(addr Address) (ValidatorInfo, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	var info ValidatorInfo
	raw, err := vm.ledger.GetState(vm.key(addr))
	if err != nil || len(raw) == 0 {
		return info, errors.New("not registered")
	}
	if err := json.Unmarshal(raw, &info); err != nil {
		return info, err
	}
	return info, nil
}

// List returns all validators. If activeOnly is true only active ones are listed.
func (vm *ValidatorManager) List(activeOnly bool) ([]ValidatorInfo, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	it := vm.ledger.PrefixIterator([]byte("validator:"))
	var out []ValidatorInfo
	for it.Next() {
		var v ValidatorInfo
		if err := json.Unmarshal(it.Value(), &v); err != nil {
			return nil, err
		}
		if activeOnly && !v.Active {
			continue
		}
		out = append(out, v)
	}
	return out, nil
}

// IsValidator checks if the address is registered and active.
func (vm *ValidatorManager) IsValidator(addr Address) bool {
	raw, err := vm.ledger.GetState(vm.key(addr))
	if err != nil || len(raw) == 0 {
		return false
	}
	var v ValidatorInfo
	_ = json.Unmarshal(raw, &v)
	return v.Active
}

func (vm *ValidatorManager) key(addr Address) []byte {
	return []byte("validator:" + addr.Hex())
}
