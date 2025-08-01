package core

// rollup_management.go - Administrative functions for controlling the roll-up aggregator.

import (
	"errors"
)

func aggregatorPausedKey() []byte { return []byte("rollup:paused") }

// PauseAggregator toggles the aggregator into a paused state. It writes the
// status to the ledger so other components can query it.
func (ag *Aggregator) PauseAggregator() error {
	ag.mu.Lock()
	defer ag.mu.Unlock()
	if ag.paused {
		return errors.New("aggregator already paused")
	}
	if err := ag.led.SetState(aggregatorPausedKey(), []byte{1}); err != nil {
		return err
	}
	ag.paused = true
	return nil
}

// ResumeAggregator lifts the pause flag and resumes normal batch submission.
func (ag *Aggregator) ResumeAggregator() error {
	ag.mu.Lock()
	defer ag.mu.Unlock()
	if !ag.paused {
		return errors.New("aggregator not paused")
	}
	if err := ag.led.SetState(aggregatorPausedKey(), []byte{0}); err != nil {
		return err
	}
	ag.paused = false
	return nil
}

// AggregatorStatus returns true if the aggregator is currently paused.
func (ag *Aggregator) AggregatorStatus() bool {
	ag.mu.Lock()
	paused := ag.paused
	ag.mu.Unlock()
	return paused
}
