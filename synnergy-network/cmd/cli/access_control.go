package cli

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var (
	acOnce sync.Once
	acCtrl *core.AccessController
)

func accessInit(cmd *cobra.Command, _ []string) error {
	var err error
	acOnce.Do(func() {
		led := core.CurrentLedger()
		if led == nil {
			err = errors.New("ledger not initialised")
			return
		}
		acCtrl = core.NewAccessController(led)
	})
	return err
}

func acDecodeAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("invalid address")
	}
	copy(a[:], b)
	return a, nil
}

func accessGrantHandler(cmd *cobra.Command, args []string) error {
	addr, err := acDecodeAddr(args[1])
	if err != nil {
		return err
	}
	return acCtrl.GrantRole(addr, args[0])
}

func accessRevokeHandler(cmd *cobra.Command, args []string) error {
	addr, err := acDecodeAddr(args[1])
	if err != nil {
		return err
	}
	return acCtrl.RevokeRole(addr, args[0])
}

func accessCheckHandler(cmd *cobra.Command, args []string) error {
	addr, err := acDecodeAddr(args[1])
	if err != nil {
		return err
	}
	if acCtrl.HasRole(addr, args[0]) {
		fmt.Fprintln(cmd.OutOrStdout(), "true")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "false")
	}
	return nil
}

func accessListHandler(cmd *cobra.Command, args []string) error {
	addr, err := acDecodeAddr(args[0])
	if err != nil {
		return err
	}
	roles, err := acCtrl.ListRoles(addr)
	if err != nil {
		return err
	}
	for _, r := range roles {
		fmt.Fprintln(cmd.OutOrStdout(), r)
	}
	return nil
}

var accessCmd = &cobra.Command{
	Use:               "access",
	Short:             "Role based access control",
	PersistentPreRunE: accessInit,
}

var acGrantCmd = &cobra.Command{Use: "grant <role> <addr>", Short: "Grant role", Args: cobra.ExactArgs(2), RunE: accessGrantHandler}
var acRevokeCmd = &cobra.Command{Use: "revoke <role> <addr>", Short: "Revoke role", Args: cobra.ExactArgs(2), RunE: accessRevokeHandler}
var acCheckCmd = &cobra.Command{Use: "check <role> <addr>", Short: "Check role", Args: cobra.ExactArgs(2), RunE: accessCheckHandler}
var acListCmd = &cobra.Command{Use: "list <addr>", Short: "List roles", Args: cobra.ExactArgs(1), RunE: accessListHandler}

func init() {
	accessCmd.AddCommand(acGrantCmd, acRevokeCmd, acCheckCmd, acListCmd)
}

// AccessCmd exports the root command.
var AccessCmd = accessCmd
