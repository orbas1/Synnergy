package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var recoveryCmd = &cobra.Command{
	Use:   "recovery",
	Short: "Account recovery operations",
}

// recoveryFlags aggregates inputs for recovery commands.
type recoveryFlags struct {
	owner    core.Address
	recovery core.Address
	phone    string
	email    string
}

var registerRecCmd = &cobra.Command{
	Use:   "register",
	Short: "Register recovery credentials",
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		ownerStr, _ := cmd.Flags().GetString("owner")
		recStr, _ := cmd.Flags().GetString("recovery")
		phone, _ := cmd.Flags().GetString("phone")
		email, _ := cmd.Flags().GetString("email")
		if ownerStr == "" || recStr == "" {
			return errors.New("--owner and --recovery required")
		}
		if phone == "" && email == "" {
			return errors.New("provide --phone or --email for contact")
		}
		owner, err := core.StringToAddress(ownerStr)
		if err != nil {
			return fmt.Errorf("invalid owner address: %w", err)
		}
		recAddr, err := core.StringToAddress(recStr)
		if err != nil {
			return fmt.Errorf("invalid recovery address: %w", err)
		}
		rf := recoveryFlags{owner: owner, recovery: recAddr, phone: phone, email: email}
		ctx := context.WithValue(cmd.Context(), "recoveryFlags", rf)
		cmd.SetContext(ctx)
		return nil
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		rf := cmd.Context().Value("recoveryFlags").(recoveryFlags)
		mgr := core.NewAccountRecovery(core.CurrentLedger())
		info := core.RecoveryInfo{RecoveryWallet: rf.recovery, PhoneNumber: rf.phone, Email: rf.email}
		return mgr.Register(rf.owner, info)
	},
}

var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Recover an account using credentials",
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		ownerStr, _ := cmd.Flags().GetString("owner")
		recStr, _ := cmd.Flags().GetString("recovery")
		phone, _ := cmd.Flags().GetString("phone")
		email, _ := cmd.Flags().GetString("email")
		if ownerStr == "" || recStr == "" {
			return errors.New("--owner and --recovery required")
		}
		if phone == "" && email == "" {
			return errors.New("provide --phone or --email for contact")
		}
		owner, err := core.StringToAddress(ownerStr)
		if err != nil {
			return fmt.Errorf("invalid owner address: %w", err)
		}
		recAddr, err := core.StringToAddress(recStr)
		if err != nil {
			return fmt.Errorf("invalid recovery address: %w", err)
		}
		rf := recoveryFlags{owner: owner, recovery: recAddr, phone: phone, email: email}
		ctx := context.WithValue(cmd.Context(), "recoveryFlags", rf)
		cmd.SetContext(ctx)
		return nil
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		rf := cmd.Context().Value("recoveryFlags").(recoveryFlags)
		mgr := core.NewAccountRecovery(core.CurrentLedger())
		info := core.RecoveryInfo{RecoveryWallet: rf.recovery, PhoneNumber: rf.phone, Email: rf.email}
		if err := mgr.Recover(rf.owner, info); err != nil {
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
	registerRecCmd.MarkFlagRequired("owner")
	registerRecCmd.MarkFlagRequired("recovery")

	recoverCmd.Flags().String("owner", "", "owner address")
	recoverCmd.Flags().String("recovery", "", "recovery wallet")
	recoverCmd.Flags().String("phone", "", "phone number")
	recoverCmd.Flags().String("email", "", "email address")
	recoverCmd.MarkFlagRequired("owner")
	recoverCmd.MarkFlagRequired("recovery")

	recoveryCmd.AddCommand(registerRecCmd, recoverCmd)
}

// RecoveryCmd exposes account recovery operations.
var RecoveryCmd = recoveryCmd

// RegisterRecovery adds recovery commands to the root CLI.
func RegisterRecovery(root *cobra.Command) { root.AddCommand(RecoveryCmd) }
