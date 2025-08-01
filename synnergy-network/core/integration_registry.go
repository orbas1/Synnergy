package core

import "sync"

// IntegrationRegistry manages external API and blockchain connections used by
// IntegrationNodes. It maintains simple in-memory indexes that can be swapped
// for persistent storage in real deployments.
type IntegrationRegistry struct {
	apis   map[string]string
	chains map[string]string
	mu     sync.RWMutex
}

// NewIntegrationRegistry creates an empty registry instance.
func NewIntegrationRegistry() *IntegrationRegistry {
	return &IntegrationRegistry{
		apis:   make(map[string]string),
		chains: make(map[string]string),
	}
}

// RegisterAPI records a reachable API endpoint under a friendly name.
func (r *IntegrationRegistry) RegisterAPI(name, endpoint string) {
	r.mu.Lock()
	r.apis[name] = endpoint
	r.mu.Unlock()
}

// RemoveAPI deletes a previously registered API.
func (r *IntegrationRegistry) RemoveAPI(name string) {
	r.mu.Lock()
	delete(r.apis, name)
	r.mu.Unlock()
}

// ListAPIs returns the names of all configured APIs.
func (r *IntegrationRegistry) ListAPIs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.apis))
	for k := range r.apis {
		out = append(out, k)
	}
	return out
}

// ConnectChain records a bridge to an external blockchain.
func (r *IntegrationRegistry) ConnectChain(id, endpoint string) {
	r.mu.Lock()
	r.chains[id] = endpoint
	r.mu.Unlock()
}

// DisconnectChain removes a chain connection from the registry.
func (r *IntegrationRegistry) DisconnectChain(id string) {
	r.mu.Lock()
	delete(r.chains, id)
	r.mu.Unlock()
}

// ListChains returns identifiers of all connected chains.
func (r *IntegrationRegistry) ListChains() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.chains))
	for id := range r.chains {
		out = append(out, id)
	}
	return out
}

// SyncData is a stub for data synchronisation routines. It can be extended
// to push or pull data between the Synthron ledger and external services.
func (r *IntegrationRegistry) SyncData(target string, payload []byte) {
	// Implementation placeholder - network IO would happen here.
	_ = target
	_ = payload
}
