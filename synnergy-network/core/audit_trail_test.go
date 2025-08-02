package core

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestAuditTrailArchive(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")
	at, err := NewAuditTrail(logPath, nil)
	if err != nil {
		t.Fatalf("NewAuditTrail: %v", err)
	}
	defer at.Close()
	if err := at.Log("event", map[string]string{"k": "v"}); err != nil {
		t.Fatalf("Log: %v", err)
	}
	outDir := filepath.Join(dir, "archive")
	if err := os.Mkdir(outDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path, sum, err := at.Archive(outDir)
	if err != nil {
		t.Fatalf("Archive: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}
	want := fmt.Sprintf("%x", sha256.Sum256(data))
	if sum != want {
		t.Fatalf("checksum mismatch: got %s want %s", sum, want)
	}
	if _, err := os.Stat(path + ".sha256"); err != nil {
		t.Fatalf("manifest missing: %v", err)
	}
}
