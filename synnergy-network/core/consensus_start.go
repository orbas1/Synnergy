package core

import "context"

// Start launches the consensus engine, spinning up proposer and block
// aggregation loops and wiring the PoS vote subscription. The method is
// idempotent and returns immediately; background routines terminate when the
// supplied context is cancelled.
func (sc *SynnergyConsensus) Start(ctx context.Context) {
	if sc == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	sc.mu.Lock()
	if sc.cancel != nil {
		sc.mu.Unlock()
		cancel()
		return
	}
	sc.cancel = cancel
	sc.mu.Unlock()

	// Launch sub‑block proposer and block sealing loops.
	go sc.subBlockLoop(ctx)
	go sc.blockLoop(ctx)

	// Subscribe to PoS vote messages if the p2p layer supports it.
	if subber, ok := sc.p2p.(interface {
		Subscribe(string) (<-chan InboundMsg, func())
	}); ok {
		sub, unsub := subber.Subscribe("posvote")
		go func() {
			defer unsub()
			for {
				select {
				case <-ctx.Done():
					return
				case m := <-sub:
					sc.handlePoSVote(m)
				}
			}
		}()
	}

	// Log lifecycle events when a logger is provided.
	if sc.logger != nil {
		sc.logger.Println("consensus started")
		go func() {
			<-ctx.Done()
			sc.logger.Println("consensus stopped")
		}()
	}
}

// Stop terminates the consensus engine started via Start. It is safe to call
// multiple times.
func (sc *SynnergyConsensus) Stop() {
	if sc == nil {
		return
	}
	sc.mu.Lock()
	cancel := sc.cancel
	sc.cancel = nil
	sc.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}
