package cli

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

// ---------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------
func ensureRegMiddleware(cmd *cobra.Command, _ []string) error {
	if core.CurrentLedger() == nil {
		return errors.New("ledger not initialised")
	}
	if core.ListRegulators() == nil {
		core.InitRegulatory(core.CurrentLedger())
	}
	return nil
}

// ---------------------------------------------------------------------
// Controller
// ---------------------------------------------------------------------

type RegController struct{}

func (r *RegController) Register(id, name, juris string) error {
	return core.RegisterRegulator(id, name, juris)
}

func (r *RegController) List() []core.Regulator { return core.ListRegulators() }

// ---------------------------------------------------------------------
// CLI commands
// ---------------------------------------------------------------------

var regCmd = &cobra.Command{
	Use:               "regulator",
	Short:             "Manage on-chain regulators",
	PersistentPreRunE: ensureRegMiddleware,
}

var regRegisterCmd = &cobra.Command{
	Use:   "register <id> <name> <jurisdiction>",
	Short: "Register a new regulator",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := &RegController{}
		if err := c.Register(args[0], args[1], args[2]); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "regulator registered")
		return nil
	},
}

var regListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all regulators",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := &RegController{}
		regs := c.List()
		b, _ := json.MarshalIndent(regs, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(b))
		return nil
	},
}

func init() {
	regCmd.AddCommand(regRegisterCmd)
	regCmd.AddCommand(regListCmd)
}

// Export
var RegulatoryCmd = regCmd
