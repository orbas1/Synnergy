package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

// helper to decode hex address
func daoParseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("invalid address")
	}
	copy(a[:], b)
	return a, nil
}

var daoCmd = &cobra.Command{
	Use:   "dao",
	Short: "Manage DAOs on the network",
}

var daoCreateCmd = &cobra.Command{
	Use:   "create <name> <creator>",
	Short: "Create a new DAO",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		creator, err := daoParseAddr(args[1])
		if err != nil {
			return err
		}
		d, err := core.CreateDAO(name, creator)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(d)
	},
}

var daoJoinCmd = &cobra.Command{
	Use:   "join <dao-id> <addr>",
	Short: "Join an existing DAO",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := daoParseAddr(args[1])
		if err != nil {
			return err
		}
		return core.JoinDAO(args[0], addr)
	},
}

var daoLeaveCmd = &cobra.Command{
	Use:   "leave <dao-id> <addr>",
	Short: "Leave a DAO",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := daoParseAddr(args[1])
		if err != nil {
			return err
		}
		return core.LeaveDAO(args[0], addr)
	},
}

var daoInfoCmd = &cobra.Command{
	Use:   "info <dao-id>",
	Short: "Show DAO details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		d, err := core.DAOInfo(args[0])
		if err != nil {
			return err
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(d)
	},
}

var daoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all DAOs",
	RunE: func(cmd *cobra.Command, args []string) error {
		ds, err := core.ListDAOs()
		if err != nil {
			return err
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(ds)
	},
}

func init() {
	daoCmd.AddCommand(daoCreateCmd, daoJoinCmd, daoLeaveCmd, daoInfoCmd, daoListCmd)
}

// Exported for index.go
var DAOCmd = daoCmd
