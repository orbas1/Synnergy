package core

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// AuditManager coordinates persistent audit logs stored on the ledger.
// Each entry can optionally be mirrored to a local AuditTrail file for
// off-chain redundancy. The manager is safe for concurrent use once
// initialised via InitAuditManager.

type AuditManager struct {
	ledger *Ledger
	trail  *AuditTrail
}

var (
	auditMgr  *AuditManager
	auditOnce sync.Once
)

// InitAuditManager creates the global audit manager. Subsequent calls are
// ignored. A non-empty file path enables on-disk logging via AuditTrail.
func InitAuditManager(ledger *Ledger, trailPath string) error {
	var initErr error
	auditOnce.Do(func() {
		var at *AuditTrail
		var err error
		if trailPath != "" {
			at, err = NewAuditTrail(trailPath, ledger)
			if err != nil {
				initErr = err
				return
			}
		}
		auditMgr = &AuditManager{ledger: ledger, trail: at}
	})
	return initErr
}

// AuditManagerInstance returns the globally configured manager.
func AuditManagerInstance() *AuditManager { return auditMgr }

// LedgerAuditEvent represents a ledger-backed audit entry.
type LedgerAuditEvent struct {
	Timestamp int64             `json:"ts"`
	Address   Address           `json:"addr"`
	Event     string            `json:"evt"`
	Meta      map[string]string `json:"meta,omitempty"`
}

func (am *AuditManager) key(addr Address, ts int64) []byte {
	b := append([]byte("auditmgr:"), addr[:]...)
	var t [8]byte
	binary.BigEndian.PutUint64(t[:], uint64(ts))
	return append(b, t[:]...)
}

// Log stores an audit event on the ledger and optionally in the local trail.
func (am *AuditManager) Log(addr Address, event string, meta map[string]string) error {
	if am == nil || am.ledger == nil {
		return errors.New("audit manager not initialised")
	}
	entry := LedgerAuditEvent{Timestamp: time.Now().Unix(), Address: addr, Event: event, Meta: meta}
	raw, _ := json.Marshal(entry)
	if err := am.ledger.SetState(am.key(addr, entry.Timestamp), raw); err != nil {
		return err
	}
	if am.trail != nil {
		return am.trail.Log(event, meta)
	}
	return nil
}

// Events fetches all audit events recorded for an address.
func (am *AuditManager) Events(addr Address) ([]LedgerAuditEvent, error) {
	if am == nil || am.ledger == nil {
		return nil, errors.New("audit manager not initialised")
	}
	prefix := append([]byte("auditmgr:"), addr[:]...)
	it := am.ledger.PrefixIterator(prefix)
	var out []LedgerAuditEvent
	for it.Next() {
		var ev LedgerAuditEvent
		if err := json.Unmarshal(it.Value(), &ev); err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	if ierr, ok := it.(interface{ Error() error }); ok {
		if err := ierr.Error(); err != nil {
			return out, err
		}
	}
	return out, nil
}

// Archive exports the audit trail to the provided destination path and
// returns the file path and sha256 checksum of the archived log.
func (am *AuditManager) Archive(dest string) (string, string, error) {
	if am == nil || am.trail == nil {
		return "", "", errors.New("audit trail not configured")
	}
	return am.trail.Archive(dest)
}

// Close closes the underlying AuditTrail if configured.
func (am *AuditManager) Close() error {
	if am == nil || am.trail == nil {
		return nil
	}
	return am.trail.Close()
}
