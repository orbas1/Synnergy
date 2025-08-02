package testutil

import (
	"io/fs"
	"os"
	"path/filepath"
)

// Sandbox provides an isolated temporary directory for tests.
type Sandbox struct {
	Root string
}

// NewSandbox creates a new Sandbox rooted at a temporary directory.
func NewSandbox() (*Sandbox, error) {
	dir, err := os.MkdirTemp("", "synnergy_sandbox")
	if err != nil {
		return nil, err
	}
	return &Sandbox{Root: dir}, nil
}

// Path returns the absolute path for a file within the sandbox.
func (s *Sandbox) Path(name string) string {
	return filepath.Join(s.Root, name)
}

// WriteFile writes data to the named file inside the sandbox using the
// provided permissions.
func (s *Sandbox) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(s.Path(name), data, perm)
}

// ReadFile reads and returns data from the named file inside the sandbox.
func (s *Sandbox) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(s.Path(name))
}

// Cleanup removes all files within the sandbox and deletes the root directory.
func (s *Sandbox) Cleanup() error {
	return os.RemoveAll(s.Root)
}
