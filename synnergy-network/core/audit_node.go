package core

import (
	"errors"
)

// AuditNode ties together a BootstrapNode with the AuditManager.
type AuditNode struct {
	node *BootstrapNode
	mgr  *AuditManager
}

// AuditNodeConfig bundles the bootstrap configuration and optional audit trail file.
type AuditNodeConfig struct {
	Bootstrap BootstrapConfig
	TrailPath string
}

// NewAuditNode initialises the underlying bootstrap node and audit manager.
func NewAuditNode(cfg *AuditNodeConfig) (*AuditNode, error) {
	if cfg == nil {
		return nil, errors.New("nil config")
	}
	b, err := NewBootstrapNode(&cfg.Bootstrap)
	if err != nil {
		return nil, err
	}
	if err := InitAuditManager(nil, cfg.TrailPath); err != nil {
		return nil, err
	}
	return &AuditNode{node: b, mgr: AuditManagerInstance()}, nil
}

// Start launches the audit node services.
func (a *AuditNode) Start() { a.node.Start() }

// Stop gracefully stops the node and closes the audit manager.
func (a *AuditNode) Stop() error {
	if err := a.node.Stop(); err != nil {
		return err
	}
	if a.mgr != nil {
		return a.mgr.Close()
	}
	return nil
}

// DialSeed proxies to the underlying network node.
func (a *AuditNode) DialSeed(peers []string) error { return a.node.DialSeed(peers) }

// Broadcast proxies to the underlying network node.
func (a *AuditNode) Broadcast(topic string, data []byte) error {
	return a.node.net.Broadcast(topic, data)
}

// Subscribe proxies to the underlying network node.
func (a *AuditNode) Subscribe(topic string) (<-chan Message, error) {
	return nil, errors.New("not implemented")
}

// ListenAndServe runs the embedded network node.
func (a *AuditNode) ListenAndServe() { a.node.net.ListenAndServe() }

// Close is an alias for Stop.
func (a *AuditNode) Close() error { return a.Stop() }

// Peers returns the current peer list.
func (a *AuditNode) Peers() []*Peer { return nil }

// LogAudit records an audit event via the manager.
func (a *AuditNode) LogAudit(addr Address, event string, meta map[string]string) error {
	if a.mgr == nil {
		return errors.New("audit manager not initialised")
	}
	return a.mgr.Log(addr, event, meta)
}

// AuditEvents retrieves past audit events for an address.
func (a *AuditNode) AuditEvents(addr Address) ([]LedgerAuditEvent, error) {
	if a.mgr == nil {
		return nil, errors.New("audit manager not initialised")
	}
	return a.mgr.Events(addr)
}

// ArchiveTrail exports the audit trail to dest and returns the created path
// and sha256 checksum of the log file.
func (a *AuditNode) ArchiveTrail(dest string) (string, string, error) {
	if a.mgr == nil {
		return "", "", errors.New("audit manager not initialised")
	}
	return a.mgr.Archive(dest)
}
