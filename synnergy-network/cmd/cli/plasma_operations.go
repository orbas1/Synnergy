package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

// ensurePlasma verifies the plasma coordinator is ready.
func ensurePlasma(_ *cobra.Command, _ []string) error {
	if core.Plasma() == nil {
		return fmt.Errorf("plasma coordinator not initialised")
	}
	return nil
}

// PlasmaController wraps core plasma helpers.
type PlasmaController struct{}

func (p *PlasmaController) Deposit(from core.Address, tok core.TokenID, amt uint64) error {
	blk := core.PlasmaBlock{Height: 0, Timestamp: time.Now().Unix()}
	_, err := core.Plasma().Deposit(from, tok, amt, blk)
	return err
}

func (p *PlasmaController) StartExit(owner core.Address, tok core.TokenID, amt uint64) error {
	blk := core.PlasmaBlock{Height: 0, Timestamp: time.Now().Unix()}
	_, err := core.Plasma().StartExit(owner, tok, amt, blk)
	return err
}

func (p *PlasmaController) Finalize(n uint64) error { return core.Plasma().FinalizeExit(n) }

//---------------------------------------------------------------------
// CLI commands
//---------------------------------------------------------------------

var plasmaOpsCmd = &cobra.Command{
	Use:               "plasmaops",
	Short:             "Interact with the plasma bridge",
	PersistentPreRunE: ensurePlasma,
}

var plasmaOpsDepositCmd = &cobra.Command{
	Use:   "deposit <from> <token> <amount>",
	Short: "Deposit tokens into the plasma bridge",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &PlasmaController{}
		from, err := decodeAddr(args[0])
		if err != nil {
			return err
		}
		token, err := decodeTokenID(args[1])
		if err != nil {
			return err
		}
		amt, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("amount uint64: %w", err)
		}
		return ctrl.Deposit(from, token, amt)
	},
}

var plasmaOpsExitCmd = &cobra.Command{
	Use:   "exit <owner> <token> <amount>",
	Short: "Start an exit from the plasma bridge",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &PlasmaController{}
		owner, err := decodeAddr(args[0])
		if err != nil {
			return err
		}
		token, err := decodeTokenID(args[1])
		if err != nil {
			return err
		}
		amt, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("amount uint64: %w", err)
		}
		return ctrl.StartExit(owner, token, amt)
	},
}

var plasmaOpsFinalizeCmd = &cobra.Command{
	Use:   "finalize <nonce>",
	Short: "Finalize a pending exit",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &PlasmaController{}
		n, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("nonce uint64: %w", err)
		}
		return ctrl.Finalize(n)
	},
}

var plasmaOpsGetCmd = &cobra.Command{
	Use:   "get <nonce>",
	Short: "Retrieve a plasma exit",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("nonce uint64: %w", err)
		}
		ex, err := core.Plasma().GetExit(n)
		if err != nil {
			return err
		}
		b, _ := json.MarshalIndent(ex, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var plasmaOpsListCmd = &cobra.Command{
	Use:   "list <owner>",
	Short: "List exits for an address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, err := decodeAddr(args[0])
		if err != nil {
			return err
		}
		list, err := core.Plasma().ListExits(owner)
		if err != nil {
			return err
		}
		b, _ := json.MarshalIndent(list, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

func init() {
	plasmaOpsCmd.AddCommand(plasmaOpsDepositCmd)
	plasmaOpsCmd.AddCommand(plasmaOpsExitCmd)
	plasmaOpsCmd.AddCommand(plasmaOpsFinalizeCmd)
	plasmaOpsCmd.AddCommand(plasmaOpsGetCmd)
	plasmaOpsCmd.AddCommand(plasmaOpsListCmd)
}

// PlasmaRoute exports the root command for registration in index.go
var PlasmaRoute = plasmaOpsCmd
