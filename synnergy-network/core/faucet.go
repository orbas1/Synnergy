package core

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// FaucetAccount is the default funding account used by the faucet.
var FaucetAccount Address

func init() {
	var err error
	FaucetAccount, err = StringToAddress("0x4661756365744163636f756e7400000000000000")
	if err != nil {
		panic("invalid FaucetAccount: " + err.Error())
	}
}

// Faucet dispenses test tokens or coins with optional rate limiting.
type Faucet struct {
	logger   *logrus.Logger
	ledger   *Ledger
	token    TokenID       // 0 means Synthron coin
	amount   uint64        // amount per request
	cooldown time.Duration // minimum time between requests per address

	mu   sync.Mutex
	last map[Address]time.Time
}

// NewFaucet creates a new faucet bound to the given ledger. The faucet
// dispenses `amount` of `token` every `cooldown` duration.
func NewFaucet(lg *logrus.Logger, led *Ledger, token TokenID, amount uint64, cooldown time.Duration) *Faucet {
	if lg == nil {
		lg = logrus.StandardLogger()
	}
	return &Faucet{
		logger:   lg,
		ledger:   led,
		token:    token,
		amount:   amount,
		cooldown: cooldown,
		last:     make(map[Address]time.Time),
	}
}

// Request sends faucet funds to the specified address if the cooldown
// period has elapsed. It returns an error if the faucet balance is
// insufficient or if rate limiting blocks the request.
func (f *Faucet) Request(to Address) error {
	if f == nil || f.ledger == nil {
		return errors.New("faucet not initialised")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now().UTC()
	if ts, ok := f.last[to]; ok {
		if now.Sub(ts) < f.cooldown {
			wait := f.cooldown - now.Sub(ts)
			return fmt.Errorf("faucet: cooldown %s remaining", wait)
		}
	}

	if f.token == 0 {
		if err := f.ledger.Transfer(FaucetAccount, to, f.amount); err != nil {
			return err
		}
	} else {
		tok, ok := GetToken(f.token)
		if !ok {
			return fmt.Errorf("token %d not found", f.token)
		}
		if err := tok.Transfer(FaucetAccount, to, f.amount); err != nil {
			return err
		}
	}

	f.last[to] = now
	f.logger.WithFields(logrus.Fields{"to": to.String(), "amount": f.amount}).Info("faucet dispense")
	return nil
}

// Balance returns the current balance held by the faucet account.
func (f *Faucet) Balance() (uint64, error) {
	if f == nil || f.ledger == nil {
		return 0, errors.New("faucet not initialised")
	}
	if f.token == 0 {
		return f.ledger.BalanceOf(FaucetAccount), nil
	}
	tok, ok := GetToken(f.token)
	if !ok {
		return 0, fmt.Errorf("token %d not found", f.token)
	}
	return tok.BalanceOf(FaucetAccount), nil
}

// SetAmount updates the amount dispensed per request.
func (f *Faucet) SetAmount(amt uint64) {
	f.mu.Lock()
	f.amount = amt
	f.mu.Unlock()
}

// SetCooldown modifies the cooldown between requests.
func (f *Faucet) SetCooldown(d time.Duration) {
	f.mu.Lock()
	f.cooldown = d
	f.mu.Unlock()
}
