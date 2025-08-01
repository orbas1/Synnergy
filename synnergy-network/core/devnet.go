package core

import (
	"errors"
	"fmt"
)

// StartDevNet spins up a number of in-memory nodes listening on sequential ports.
// It returns the running nodes so the caller can manage their lifecycle.
func StartDevNet(nodes int) ([]*Node, error) {
	if nodes <= 0 {
		return nil, errors.New("number of nodes must be positive")
	}
	list := make([]*Node, nodes)
	for i := 0; i < nodes; i++ {
		cfg := Config{
			ListenAddr:     fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 4101+i),
			BootstrapPeers: []string{},
			DiscoveryTag:   fmt.Sprintf("devnet-%d", i),
		}
		n, err := NewNode(cfg)
		if err != nil {
			for j := 0; j < i; j++ {
				_ = list[j].Close()
			}
			return nil, fmt.Errorf("start devnet: %w", err)
		}
		list[i] = n
		go n.ListenAndServe()
	}
	return list, nil
}

// StartTestNet creates nodes from explicit configurations. Each node is started
// in its own goroutine and returned for management by the caller.
func StartTestNet(cfgs []Config) ([]*Node, error) {
	if len(cfgs) == 0 {
		return nil, errors.New("no node configurations supplied")
	}
	nodes := make([]*Node, len(cfgs))
	for i, cfg := range cfgs {
		n, err := NewNode(cfg)
		if err != nil {
			for j := 0; j < i; j++ {
				_ = nodes[j].Close()
			}
			return nil, fmt.Errorf("start testnet: %w", err)
		}
		nodes[i] = n
		go n.ListenAndServe()
	}
	return nodes, nil
}
