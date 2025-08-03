//go:build !tokens

package core

import "context"

// Start is a no-op placeholder for builds that exclude the full consensus
// engine. It satisfies interfaces expecting a Start method.
func (sc *SynnergyConsensus) Start(ctx context.Context) {
	// Intentionally left blank.
}

// Stop is a no-op placeholder for builds that exclude the full consensus
// engine. It allows callers to compile without the full implementation.
func (sc *SynnergyConsensus) Stop() {
	// Intentionally left blank.
}
