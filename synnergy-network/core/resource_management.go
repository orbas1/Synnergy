package core

import (
	"encoding/json"
	"fmt"
	"sync"
)

// ResourceQuota tracks allowed and consumed resources for an address.
type ResourceQuota struct {
	CPU     uint64 `json:"cpu"`
	Memory  uint64 `json:"memory"`
	Storage uint64 `json:"storage"`
	UsedCPU uint64 `json:"used_cpu"`
	UsedMem uint64 `json:"used_mem"`
	UsedSto uint64 `json:"used_storage"`
}

type ResourceManager struct {
	led StateRW
	mu  sync.Mutex
}

var (
	rmOnce sync.Once
	rm     *ResourceManager
)

// InitResourceManager initialises the global resource manager.
func InitResourceManager(led StateRW) { rmOnce.Do(func() { rm = &ResourceManager{led: led} }) }

// RM returns the global resource manager instance.
func RM() *ResourceManager { return rm }

func quotaKey(addr Address) []byte { return append([]byte("quota:"), addr.Bytes()...) }

// SetQuota configures resource limits for an address.
func (m *ResourceManager) SetQuota(addr Address, cpu, mem, store uint64) error {
	q := ResourceQuota{CPU: cpu, Memory: mem, Storage: store}
	b, err := json.Marshal(q)
	if err != nil {
		return err
	}
	return m.led.SetState(quotaKey(addr), b)
}

// GetQuota retrieves quota information for an address.
func (m *ResourceManager) GetQuota(addr Address) (ResourceQuota, error) {
	b, err := m.led.GetState(quotaKey(addr))
	if err != nil {
		return ResourceQuota{}, err
	}
	var q ResourceQuota
	if err := json.Unmarshal(b, &q); err != nil {
		return ResourceQuota{}, err
	}
	return q, nil
}

// ChargeResources records usage and deducts payment from the address.
func (m *ResourceManager) ChargeResources(addr Address, cpu, mem, store uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	q, err := m.GetQuota(addr)
	if err != nil {
		return fmt.Errorf("quota read: %w", err)
	}
	if q.UsedCPU+cpu > q.CPU || q.UsedMem+mem > q.Memory || q.UsedSto+store > q.Storage {
		return fmt.Errorf("quota exceeded")
	}
	coin := Coin{ledger: CurrentLedger()}
	if err := coin.Transfer(addr.Bytes(), AddressZero.Bytes(), cpu+mem+store); err != nil {
		return fmt.Errorf("payment: %w", err)
	}
	q.UsedCPU += cpu
	q.UsedMem += mem
	q.UsedSto += store
	b, _ := json.Marshal(q)
	return m.led.SetState(quotaKey(addr), b)
}

// ReleaseResources reduces recorded usage without refund.
func (m *ResourceManager) ReleaseResources(addr Address, cpu, mem, store uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	q, err := m.GetQuota(addr)
	if err != nil {
		return fmt.Errorf("quota read: %w", err)
	}
	if cpu > q.UsedCPU || mem > q.UsedMem || store > q.UsedSto {
		return fmt.Errorf("release exceeds usage")
	}
	q.UsedCPU -= cpu
	q.UsedMem -= mem
	q.UsedSto -= store
	b, _ := json.Marshal(q)
	return m.led.SetState(quotaKey(addr), b)
}

//---------------------------------------------------------------------
// END resource_management.go
//---------------------------------------------------------------------
