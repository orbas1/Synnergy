package core

// ResourceAllocator tracks per-address gas allowances.
type ResourceAllocator struct {
	limits map[Address]uint64
}

// NewResourceAllocator creates a new allocator instance.
func NewResourceAllocator() *ResourceAllocator {
	return &ResourceAllocator{limits: make(map[Address]uint64)}
}

// Adjust sets the gas limit for the given address.
func (r *ResourceAllocator) Adjust(addr Address, gas uint64) {
	if r == nil {
		return
	}
	r.limits[addr] = gas
}
