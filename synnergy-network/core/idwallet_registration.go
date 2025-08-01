package core

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// IDRegistry manages on-chain registration of wallets that
// receive the SYN-ID token. A single instance is initialised
// via InitIDRegistry and reused by the CLI and VM.

type IDRegistry struct {
	led    *Ledger
	logger *logrus.Logger
}

type IDRegistration struct {
	Address Address `json:"address"`
	Details string  `json:"details"`
}

var (
	idRegOnce sync.Once
	idReg     *IDRegistry
)

// InitIDRegistry wires the ledger and logger. It must be called
// before invoking RegisterIDWallet or IsIDWalletRegistered.
func InitIDRegistry(lg *logrus.Logger, led *Ledger) {
	idRegOnce.Do(func() {
		idReg = &IDRegistry{led: led, logger: lg}
	})
}

// RegisterIDWallet stores registration details for the address and
// mints one SYN-ID governance token to it. An error is returned if the
// wallet was already registered.
func RegisterIDWallet(addr Address, info string) error {
	if idReg == nil {
		return fmt.Errorf("id registry not initialised")
	}
	key := append([]byte("idwallet:"), addr[:]...)
	if ok, _ := idReg.led.HasState(key); ok {
		return fmt.Errorf("wallet already registered")
	}
	rec := IDRegistration{Address: addr, Details: info}
	data, _ := json.Marshal(rec)
	if err := idReg.led.SetState(key, data); err != nil {
		return err
	}
	if err := idReg.led.MintToken(addr, "SYN-ID", 1); err != nil {
		return err
	}
	if idReg.logger != nil {
		idReg.logger.WithField("addr", addr.Hex()).Info("id wallet registered")
	}
	return nil
}

// IsIDWalletRegistered returns true if the address previously
// called RegisterIDWallet.
func IsIDWalletRegistered(addr Address) bool {
	if idReg == nil {
		return false
	}
	key := append([]byte("idwallet:"), addr[:]...)
	ok, _ := idReg.led.HasState(key)
	return ok
}
