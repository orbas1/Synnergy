// Package utils provides shared utility helpers used across Synnergy.
// See Version for the module's semantic version.
package utils

import "fmt"

// Wrap adds context to an error message. It returns nil if err is nil.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}
