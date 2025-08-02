package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

// helper to parse address from hex
func escParseAddr(hexStr string) (core.Address, error) {
	return core.ParseAddress(hexStr)
}

// ---------------------------- Controllers ----------------------------

type EscrowController struct{}

func (EscrowController) Create(parties []core.EscrowParty) (*core.EscrowContract, error) {
	return core.EscrowCreate(&core.Context{}, parties)
}

func (EscrowController) Deposit(id string, amt uint64) error {
	return core.EscrowDeposit(&core.Context{}, id, amt)
}

func (EscrowController) Release(id string) error                     { return core.EscrowRelease(&core.Context{}, id) }
func (EscrowController) Cancel(id string) error                      { return core.EscrowCancel(&core.Context{}, id) }
func (EscrowController) Get(id string) (*core.EscrowContract, error) { return core.EscrowGet(id) }
func (EscrowController) List() ([]core.EscrowContract, error)        { return core.EscrowList() }

// ------------------------------ CLI ---------------------------------

var escrowCmd = &cobra.Command{
	Use:   "escrow",
	Short: "Manage multi-party escrows",
}

var escrowCreateCmd = &cobra.Command{
	Use:   "create <amount> <addr1> [addrN...]",
	Short: "Create a new escrow splitting amount equally",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		total, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid amount: %w", err)
		}
		n := len(args) - 1
		share := total / uint64(n)
		parties := make([]core.EscrowParty, n)
		for i, hexAddr := range args[1:] {
			addr, err := escParseAddr(hexAddr)
			if err != nil {
				return err
			}
			parties[i] = core.EscrowParty{Address: addr, Amount: share}
		}
		esc, err := EscrowController{}.Create(parties)
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(esc, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(out))
		return nil
	},
}

var escrowDepositCmd = &cobra.Command{
	Use:   "deposit <escrow-id> <amount>",
	Short: "Deposit additional funds to an escrow",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		amt, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return err
		}
		return EscrowController{}.Deposit(args[0], amt)
	},
}

var escrowReleaseCmd = &cobra.Command{
	Use:   "release <escrow-id>",
	Short: "Release funds to escrow parties",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return EscrowController{}.Release(args[0])
	},
}

var escrowCancelCmd = &cobra.Command{
	Use:   "cancel <escrow-id>",
	Short: "Cancel an escrow and refund creator",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return EscrowController{}.Cancel(args[0])
	},
}

var escrowInfoCmd = &cobra.Command{
	Use:   "info <escrow-id>",
	Short: "Show escrow details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		esc, err := EscrowController{}.Get(args[0])
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(esc, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(out))
		return nil
	},
}

var escrowListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all escrows",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		escs, err := EscrowController{}.List()
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(escs, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(out))
		return nil
	},
}

func init() {
	escrowCmd.AddCommand(escrowCreateCmd)
	escrowCmd.AddCommand(escrowDepositCmd)
	escrowCmd.AddCommand(escrowReleaseCmd)
	escrowCmd.AddCommand(escrowCancelCmd)
	escrowCmd.AddCommand(escrowInfoCmd)
	escrowCmd.AddCommand(escrowListCmd)
}

// EscrowRoute is exported for registration in the main CLI.
var EscrowRoute = escrowCmd
