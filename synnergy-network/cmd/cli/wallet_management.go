package cli

import (
	"fmt"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var (
	wmLedger *core.Ledger
	wm       *core.WalletManager
)

func wmInit(cmd *cobra.Command, _ []string) error {
	if wm != nil {
		return nil
	}
	_ = godotenv.Load()
	var err error
	wmLedger, err = core.NewLedger(core.LedgerConfig{})
	if err != nil {
		return err
	}
	wm = core.NewWalletManager(wmLedger)
	return nil
}

var wmCmd = &cobra.Command{
	Use:               "wallet_mgmt",
	Short:             "High level wallet management",
	PersistentPreRunE: wmInit,
}

var wmCreateCmd = &cobra.Command{
	Use:   "create [bits]",
	Short: "Create a new wallet",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bits, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		_, mnemonic, err := wm.Create(bits)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), mnemonic)
		return nil
	},
}

var wmBalanceCmd = &cobra.Command{
	Use:   "balance [addr]",
	Short: "Show balance for address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := core.ParseAddress(args[0])
		if err != nil {
			return err
		}
		bal := wm.Balance(addr)
		fmt.Fprintln(cmd.OutOrStdout(), bal)
		return nil
	},
}

var wmTransferCmd = &cobra.Command{
	Use:   "transfer [mnemonic] [to] [amt]",
	Short: "Transfer SYNN using mnemonic",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		w, err := core.WalletFromMnemonic(args[0], "")
		if err != nil {
			return err
		}
		to, err := core.ParseAddress(args[1])
		if err != nil {
			return err
		}
		amt, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return err
		}
		tx, err := wm.Transfer(w, 0, 0, to, amt, 0)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%x\n", tx.Hash)
		return nil
	},
}

func init() {
	wmCmd.AddCommand(wmCreateCmd, wmBalanceCmd, wmTransferCmd)
}

var WalletMgmtCmd = wmCmd

func RegisterWalletMgmt(root *cobra.Command) { root.AddCommand(WalletMgmtCmd) }
