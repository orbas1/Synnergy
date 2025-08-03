package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"synnergy-network/core"
)

// -----------------------------------------------------------------------------
// plasma_management.go - CLI wrappers for the Plasma manager
// -----------------------------------------------------------------------------

var plasmamgmtCmd = &cobra.Command{
	Use:   "plasmamgmt",
	Short: "Manage the plasma chain",
}

var plasmamgmtDepositCmd = &cobra.Command{
	Use:   "deposit [from] [tokenID] [amount]",
	Short: "Deposit tokens into the plasma chain",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		from, err := hex.DecodeString(args[0])
		if err != nil || len(from) != len(core.AddressZero) {
			return fmt.Errorf("invalid from address")
		}
		var addr core.Address
		copy(addr[:], from)
		tokID, err := strconv.ParseUint(args[1], 10, 32)
		if err != nil {
			return err
		}
		amount, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return err
		}
		dep, err := core.Plasma().Deposit(addr, core.TokenID(tokID), amount)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "deposit id %d\n", dep.ID)
		return nil
	},
}

var plasmamgmtWithdrawCmd = &cobra.Command{
	Use:   "withdraw [depositID] [to]",
	Short: "Withdraw a plasma deposit",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return err
		}
		toBytes, err := hex.DecodeString(args[1])
		if err != nil || len(toBytes) != len(core.AddressZero) {
			return fmt.Errorf("invalid address")
		}
		var to core.Address
		copy(to[:], toBytes)
		return core.Plasma().Withdraw(id, to)
	},
}

var plasmamgmtSubmitCmd = &cobra.Command{
	Use:   "submit [root]",
	Short: "Submit a plasma block commitment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := hex.DecodeString(args[0])
		if err != nil {
			return err
		}
		blk, err := core.Plasma().SubmitBlock(root)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "block %d stored\n", blk.Height)
		return nil
	},
}

func init() {
	plasmamgmtCmd.AddCommand(plasmamgmtDepositCmd)
	plasmamgmtCmd.AddCommand(plasmamgmtWithdrawCmd)
	plasmamgmtCmd.AddCommand(plasmamgmtSubmitCmd)
}

// PlasmaMgmtCmd is the exported command to mount on the root CLI.
var PlasmaMgmtCmd = plasmamgmtCmd
