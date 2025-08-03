package core

import "context"

// Start initializes consensus processing loops.
// It currently logs startup and listens for context cancellation.
func (sc *SynnergyConsensus) Start(ctx context.Context) {
	if sc == nil {
		return
	}
	if sc.logger != nil {
		sc.logger.Println("consensus started")
	}
	go func() {
		<-ctx.Done()
		if sc.logger != nil {
			sc.logger.Println("consensus stopped")
		}
	}()
}
