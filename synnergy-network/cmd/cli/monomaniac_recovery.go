package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var recoveryCmd = &cobra.Command{
	Use:   "recovery",
	Short: "Account recovery operations",
}

var registerRecCmd = &cobra.Command{
	Use:   "register",
	Short: "Register recovery credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		ownerStr, _ := cmd.Flags().GetString("owner")
		recStr, _ := cmd.Flags().GetString("recovery")
		phone, _ := cmd.Flags().GetString("phone")
		email, _ := cmd.Flags().GetString("email")
		if ownerStr == "" {
			return errors.New("--owner required")
		}
		owner, err := core.StringToAddress(ownerStr)
		if err != nil {
			return err
		}
		recAddr, _ := core.StringToAddress(recStr)
		mgr := core.NewAccountRecovery(core.CurrentLedger())
		info := core.RecoveryInfo{RecoveryWallet: recAddr, PhoneNumber: phone, Email: email}
		return mgr.Register(owner, info)
	},
}

var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Recover an account using credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		ownerStr, _ := cmd.Flags().GetString("owner")
		recStr, _ := cmd.Flags().GetString("recovery")
		phone, _ := cmd.Flags().GetString("phone")
		email, _ := cmd.Flags().GetString("email")
		if ownerStr == "" {
			return errors.New("--owner required")
		}
		owner, err := core.StringToAddress(ownerStr)
		if err != nil {
			return err
		}
		recAddr, _ := core.StringToAddress(recStr)
		mgr := core.NewAccountRecovery(core.CurrentLedger())
		info := core.RecoveryInfo{RecoveryWallet: recAddr, PhoneNumber: phone, Email: email}
		if err := mgr.Recover(owner, info); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "account recovered")
		return nil
	},
}

func init() {
	registerRecCmd.Flags().String("owner", "", "owner address")
	registerRecCmd.Flags().String("recovery", "", "recovery wallet")
	registerRecCmd.Flags().String("phone", "", "phone number")
	registerRecCmd.Flags().String("email", "", "email address")
	recoverCmd.Flags().String("owner", "", "owner address")
	recoverCmd.Flags().String("recovery", "", "recovery wallet")
	recoverCmd.Flags().String("phone", "", "phone number")
	recoverCmd.Flags().String("email", "", "email address")
	recoveryCmd.AddCommand(registerRecCmd, recoverCmd)
}

var RecoveryCmd = recoveryCmd

func RegisterRecovery(root *cobra.Command) { root.AddCommand(RecoveryCmd) }
