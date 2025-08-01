package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"synnergy-network/core"
)

var plasmaCmd = &cobra.Command{
	Use:   "plasma",
	Short: "Interact with the Plasma chain",
}

var plasmaDepositCmd = &cobra.Command{
	Use:   "deposit",
	Short: "Deposit funds into Plasma",
	RunE: func(cmd *cobra.Command, args []string) error {
		fromStr, _ := cmd.Flags().GetString("from")
		amt, _ := cmd.Flags().GetUint64("amt")
		from, err := parseAddr(fromStr)
		if err != nil {
			return err
		}
		_, err = core.Plasma().Deposit(from, amt)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "deposit ok ✔")
		return nil
	},
}

var plasmaWithdrawCmd = &cobra.Command{
	Use:   "withdraw",
	Short: "Withdraw from Plasma by nonce",
	RunE: func(cmd *cobra.Command, args []string) error {
		nonce, _ := cmd.Flags().GetUint64("nonce")
		toStr, _ := cmd.Flags().GetString("to")
		to, err := parseAddr(toStr)
		if err != nil {
			return err
		}
		err = core.Plasma().Withdraw(core.PlasmaExit{Nonce: nonce, To: to})
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "withdraw ok ✔")
		return nil
	},
}

func init() {
	plasmaDepositCmd.Flags().String("from", "", "sender address")
	plasmaDepositCmd.Flags().Uint64("amt", 0, "amount to deposit")
	plasmaWithdrawCmd.Flags().Uint64("nonce", 0, "deposit nonce")
	plasmaWithdrawCmd.Flags().String("to", "", "recipient address")
	plasmaCmd.AddCommand(plasmaDepositCmd, plasmaWithdrawCmd)
}

// PlasmaCmd exports the root command.
var PlasmaCmd = plasmaCmd
