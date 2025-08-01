package cli

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"synnergy-network/core"
)

var (
	idLedger *core.Ledger
	idLogger = logrus.New()
	idOnce   sync.Once
)

func idMiddleware(cmd *cobra.Command, _ []string) error {
	var err error
	idOnce.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = errors.New("LEDGER_PATH not set")
			return
		}
		idLogger.SetLevel(logrus.InfoLevel)
		idLedger, err = core.OpenLedger(path)
		if err != nil {
			return
		}
		core.InitIDRegistry(idLogger, idLedger)
	})
	return err
}

func idParseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

var idwalletCmd = &cobra.Command{
	Use:               "idwallet",
	Short:             "Register wallets for SYN-ID governance",
	PersistentPreRunE: idMiddleware,
}

var idRegisterCmd = &cobra.Command{
	Use:   "register <address> <info>",
	Short: "Register wallet and mint SYN-ID token",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := idParseAddr(args[0])
		if err != nil {
			return err
		}
		info := args[1]
		return core.RegisterIDWallet(addr, info)
	},
}

var idCheckCmd = &cobra.Command{
	Use:   "check <address>",
	Short: "Check registration status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := idParseAddr(args[0])
		if err != nil {
			return err
		}
		ok := core.IsIDWalletRegistered(addr)
		fmt.Printf("registered: %v\n", ok)
		return nil
	},
}

func init() {
	idwalletCmd.AddCommand(idRegisterCmd)
	idwalletCmd.AddCommand(idCheckCmd)
}

var IDWalletCmd = idwalletCmd
