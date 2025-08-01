package cli

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

// Controller provides thin wrappers around the core CCSN module.
type CCSNController struct{}

func (c *CCSNController) Register(src, tgt string) (core.CCSNetwork, error) {
	n := core.CCSNetwork{SourceConsensus: src, TargetConsensus: tgt}
	if err := core.RegisterCCSNetwork(n); err != nil {
		return core.CCSNetwork{}, err
	}
	return n, nil
}

func (c *CCSNController) List() ([]core.CCSNetwork, error)       { return core.ListCCSNetworks() }
func (c *CCSNController) Get(id string) (core.CCSNetwork, error) { return core.GetCCSNetwork(id) }

//-------------------------------------------------------------------------
// CLI declarations
//-------------------------------------------------------------------------

var ccsnCmd = &cobra.Command{
	Use:   "ccsn",
	Short: "Manage cross‑consensus scaling networks",
}

var ccsnRegisterCmd = &cobra.Command{
	Use:   "register <source_consensus> <target_consensus>",
	Short: "Register a cross‑consensus network",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &CCSNController{}
		n, err := ctrl.Register(args[0], args[1])
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(n, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

var ccsnListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured CCS networks",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctrl := &CCSNController{}
		nets, err := ctrl.List()
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(nets, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

var ccsnGetCmd = &cobra.Command{
	Use:   "get <network_id>",
	Short: "Retrieve a CCS network by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("invalid UUID: %w", err)
		}
		ctrl := &CCSNController{}
		n, err := ctrl.Get(id)
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(n, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

func init() {
	ccsnCmd.AddCommand(ccsnRegisterCmd)
	ccsnCmd.AddCommand(ccsnListCmd)
	ccsnCmd.AddCommand(ccsnGetCmd)
}

// Export for root command
var CCSNCmd = ccsnCmd
