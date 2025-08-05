package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Manage simple accounts and balances",
}

func acctHandleCreate(cmd *cobra.Command, args []string) error {
	addr, err := core.StringToAddress(args[0])
	if err != nil {
		return err
	}
	am := core.NewAccountManager(core.CurrentLedger())
	if err := am.CreateAccount(addr); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "account created")
	return nil
}

func acctHandleDelete(cmd *cobra.Command, args []string) error {
	addr, err := core.StringToAddress(args[0])
	if err != nil {
		return err
	}
	am := core.NewAccountManager(core.CurrentLedger())
	if err := am.DeleteAccount(addr); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "account deleted")
	return nil
}

func acctHandleBalance(cmd *cobra.Command, args []string) error {
	addr, err := core.StringToAddress(args[0])
	if err != nil {
		return err
	}
	am := core.NewAccountManager(core.CurrentLedger())
	bal, err := am.Balance(addr)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%d\n", bal)
	return nil
}

func acctHandleTransfer(cmd *cobra.Command, args []string) error {
	fromStr, err := cmd.Flags().GetString("from")
	if err != nil {
		return err
	}
	toStr, err := cmd.Flags().GetString("to")
	if err != nil {
		return err
	}
	amt, err := cmd.Flags().GetUint64("amt")
	if err != nil {
		return err
	}
	from, err := core.StringToAddress(fromStr)
	if err != nil {
		return err
	}
	to, err := core.StringToAddress(toStr)
	if err != nil {
		return err
	}
	am := core.NewAccountManager(core.CurrentLedger())
	if err := am.Transfer(from, to, amt); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "transfer ok")
	return nil
}

var acctCreateCmd = &cobra.Command{Use: "create <addr>", Short: "Create account", Args: cobra.ExactArgs(1), RunE: acctHandleCreate}
var acctDeleteCmd = &cobra.Command{Use: "delete <addr>", Short: "Delete account", Args: cobra.ExactArgs(1), RunE: acctHandleDelete}
var acctBalanceCmd = &cobra.Command{Use: "balance <addr>", Short: "Show balance", Args: cobra.ExactArgs(1), RunE: acctHandleBalance}
var acctTransferCmd = &cobra.Command{Use: "transfer", Short: "Transfer between accounts", Args: cobra.NoArgs, RunE: acctHandleTransfer}

func init() {
	acctTransferCmd.Flags().String("from", "", "sender")
	acctTransferCmd.Flags().String("to", "", "recipient")
	acctTransferCmd.Flags().Uint64("amt", 0, "amount")
	acctTransferCmd.MarkFlagRequired("from")
	acctTransferCmd.MarkFlagRequired("to")
	acctTransferCmd.MarkFlagRequired("amt")

	accountCmd.AddCommand(acctCreateCmd, acctDeleteCmd, acctBalanceCmd, acctTransferCmd)
}

// AccountCmd exposes account management subcommands.
var AccountCmd = accountCmd
