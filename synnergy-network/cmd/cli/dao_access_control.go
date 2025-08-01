package cli

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core"
)

var daoCtrl *core.DAOAccessControl

func initDAOMiddleware(cmd *cobra.Command, _ []string) error {
	if daoCtrl != nil {
		return nil
	}
	lp := viper.GetString("LEDGER_PATH")
	if lp == "" {
		return errors.New("LEDGER_PATH not set")
	}
	if err := core.InitLedger(lp); err != nil {
		return err
	}
	daoCtrl = core.NewDAOAccessControl(core.CurrentLedger())
	return nil
}

var daoCmd = &cobra.Command{
	Use:               "dao_access",
	Short:             "Manage DAO access control",
	PersistentPreRunE: initDAOMiddleware,
}

var daoAddCmd = &cobra.Command{
	Use:   "add <address> <role>",
	Short: "Add a DAO member",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		b, err := hex.DecodeString(strings.TrimPrefix(args[0], "0x"))
		if err != nil || len(b) != 20 {
			return fmt.Errorf("invalid address")
		}
		var addr core.Address
		copy(addr[:], b)
		role := core.DAORoleMember
		if strings.ToLower(args[1]) == "admin" {
			role = core.DAORoleAdmin
		}
		return daoCtrl.AddMember(addr, role)
	},
}

var daoRemoveCmd = &cobra.Command{
	Use:   "remove <address>",
	Short: "Remove a DAO member",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		b, err := hex.DecodeString(strings.TrimPrefix(args[0], "0x"))
		if err != nil || len(b) != 20 {
			return fmt.Errorf("invalid address")
		}
		var addr core.Address
		copy(addr[:], b)
		return daoCtrl.RemoveMember(addr)
	},
}

var daoRoleCmd = &cobra.Command{
	Use:   "role <address>",
	Short: "Show DAO role of an address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		b, err := hex.DecodeString(strings.TrimPrefix(args[0], "0x"))
		if err != nil || len(b) != 20 {
			return fmt.Errorf("invalid address")
		}
		var addr core.Address
		copy(addr[:], b)
		role, err := daoCtrl.RoleOf(addr)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", role)
		return nil
	},
}

var daoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List DAO members",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		members, err := daoCtrl.ListMembers(0)
		if err != nil {
			return err
		}
		for _, m := range members {
			fmt.Printf("%x %s\n", m.Addr[:], m.Role)
		}
		return nil
	},
}

func init() {
	daoCmd.AddCommand(daoAddCmd, daoRemoveCmd, daoRoleCmd, daoListCmd)
}

// DAOAccessCmd is the exported command for index.go
var DAOAccessCmd = daoCmd
