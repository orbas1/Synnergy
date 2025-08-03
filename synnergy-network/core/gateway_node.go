package core

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	Nodes "synnergy-network/core/Nodes"
)

// GatewayConfig bundles dependencies required for a GatewayNode.
type GatewayConfig struct {
	Node   Nodes.NodeInterface
	Ledger *Ledger
}

// GatewayNode provides cross-chain connectivity and external data integration.
type GatewayNode struct {
	Nodes.NodeInterface
	ledger      *Ledger
	mu          sync.RWMutex
	connections map[string]ChainConnection
	externals   map[string]string
}

// NewGatewayNode creates a GatewayNode from an existing network node and ledger.
func NewGatewayNode(cfg GatewayConfig) *GatewayNode {
	g := &GatewayNode{
		NodeInterface: cfg.Node,
		ledger:        cfg.Ledger,
		connections:   make(map[string]ChainConnection),
		externals:     make(map[string]string),
	}
	return g
}

// ConnectChain establishes a new cross-chain connection via the ledger.
func (g *GatewayNode) ConnectChain(local, remote string) (ChainConnection, error) {
	conn, err := OpenChainConnection(local, remote)
	if err != nil {
		return ChainConnection{}, err
	}
	g.mu.Lock()
	g.connections[conn.ID] = conn
	g.mu.Unlock()
	return conn, nil
}

// DisconnectChain closes an existing connection.
func (g *GatewayNode) DisconnectChain(id string) error {
	if err := CloseChainConnection(id); err != nil {
		return err
	}
	g.mu.Lock()
	delete(g.connections, id)
	g.mu.Unlock()
	return nil
}

// ListConnections returns active cross-chain links tracked by this node.
func (g *GatewayNode) ListConnections() []ChainConnection {
	g.mu.RLock()
	out := make([]ChainConnection, 0, len(g.connections))
	for _, c := range g.connections {
		out = append(out, c)
	}
	g.mu.RUnlock()
	return out
}

// RegisterExternalSource registers an HTTP endpoint for data ingestion.
func (g *GatewayNode) RegisterExternalSource(name, url string) {
	g.mu.Lock()
	g.externals[name] = url
	g.mu.Unlock()
}

// RemoveExternalSource removes a previously added data source.
func (g *GatewayNode) RemoveExternalSource(name string) {
	g.mu.Lock()
	delete(g.externals, name)
	g.mu.Unlock()
}

// ExternalSources returns a copy of the configured sources.
func (g *GatewayNode) ExternalSources() map[string]string {
	g.mu.RLock()
	out := make(map[string]string, len(g.externals))
	for k, v := range g.externals {
		out[k] = v
	}
	g.mu.RUnlock()
	return out
}

// QueryExternalData fetches data from a registered HTTP source.
func (g *GatewayNode) QueryExternalData(name string) ([]byte, error) {
	g.mu.RLock()
	url, ok := g.externals[name]
	g.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown source %s", name)
	}
	resp, err := http.Get(url)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// PushExternalData publishes external payloads to the network.
func (g *GatewayNode) PushExternalData(name string, data []byte) error {
	topic := "gateway:" + name
	return g.Broadcast(topic, data)
}
