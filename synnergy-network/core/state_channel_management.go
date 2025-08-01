package core

// Enterprise-grade state channel management operations.
// Provides pause/resume controls and immediate settlement helpers.

import (
	"errors"
)

// PauseChannel marks a channel as paused preventing further updates
// until resumed. The channel must not be closing.
func (e *ChannelEngine) PauseChannel(id ChannelID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ch, err := e.getChannel(id)
	if err != nil {
		return err
	}
	if ch.Paused {
		return errors.New("already paused")
	}
	if ch.Closing != 0 {
		return errors.New("cannot pause closing channel")
	}
	ch.Paused = true
	return e.led.SetState(chKey(id), mustJSON(ch))
}

// ResumeChannel re-enables updates for a previously paused channel.
func (e *ChannelEngine) ResumeChannel(id ChannelID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ch, err := e.getChannel(id)
	if err != nil {
		return err
	}
	if !ch.Paused {
		return errors.New("not paused")
	}
	ch.Paused = false
	return e.led.SetState(chKey(id), mustJSON(ch))
}

// CancelClose aborts the pending close operation if it is still within
// the challenge period.
func (e *ChannelEngine) CancelClose(id ChannelID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ch, err := e.getChannel(id)
	if err != nil {
		return err
	}
	if ch.Closing == 0 {
		return errors.New("not closing")
	}

	ch.Closing = 0
	if err := e.led.SetState(chKey(id), mustJSON(ch)); err != nil {
		return err
	}
	e.led.DeleteState(pendingKey(id))
	return nil
}

// ForceClose immediately settles the channel using the provided signed
// state. Both parties must sign off on this final state.
func (e *ChannelEngine) ForceClose(state SignedState) error {
	if err := verifySigs(&state); err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	cur, err := e.getChannel(state.Channel.ID)
	if err != nil {
		return err
	}
	if state.Channel.Nonce < cur.Nonce {
		return errors.New("nonce too low")
	}

	tok, _ := GetToken(state.Channel.Token)
	escrow := escrowAddr(state.Channel.ID)
	if state.Channel.BalanceA > 0 {
		_ = tok.Transfer(escrow, state.Channel.PartyA, state.Channel.BalanceA)
	}
	if state.Channel.BalanceB > 0 {
		_ = tok.Transfer(escrow, state.Channel.PartyB, state.Channel.BalanceB)
	}

	e.led.DeleteState(chKey(state.Channel.ID))
	e.led.DeleteState(pendingKey(state.Channel.ID))
	return nil
}
