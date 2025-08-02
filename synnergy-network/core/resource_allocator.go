package core

import "sync"

// ResourceAllocator tracks per-address gas limits for contracts or accounts.
type ResourceAllocator struct {
	mu     sync.Mutex
	limits map[Address]uint64
}

// NewResourceAllocator creates a new allocator instance.
func NewResourceAllocator() *ResourceAllocator {
	return &ResourceAllocator{limits: make(map[Address]uint64)}
}

// Adjust sets the gas limit for addr to gas.
func (r *ResourceAllocator) Adjust(addr Address, gas uint64) {
	r.mu.Lock()
	r.limits[addr] = gas
	r.mu.Unlock()
}
