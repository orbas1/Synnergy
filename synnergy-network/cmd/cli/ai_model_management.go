package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

// AIModelController wraps management helpers around the core package.
type AIModelController struct{}

func (c *AIModelController) Get(id string) (core.ModelListing, error) {
	return core.GetModelListing(id)
}

func (c *AIModelController) List() ([]core.ModelListing, error) {
	return core.ListModelListings()
}

func (c *AIModelController) Update(id string, price uint64) error {
	seller := core.ModuleAddress("cli")
	return core.UpdateListingPrice(id, seller, price)
}

func (c *AIModelController) Remove(id string) error {
	seller := core.ModuleAddress("cli")
	return core.RemoveListing(id, seller)
}

var aiMgmtCmd = &cobra.Command{Use: "ai_mgmt", Short: "Manage AI model listings"}

var mgmtGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get listing metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AIModelController{}
		m, err := ctrl.Get(args[0])
		if err != nil {
			return err
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(m)
	},
}

var mgmtListCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all model listings",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AIModelController{}
		list, err := ctrl.List()
		if err != nil {
			return err
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(list)
	},
}

var mgmtUpdateCmd = &cobra.Command{
	Use:   "update <id> <price>",
	Short: "Update listing price",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid price: %w", err)
		}
		ctrl := &AIModelController{}
		return ctrl.Update(args[0], p)
	},
}

var mgmtRemoveCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove a listing",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AIModelController{}
		return ctrl.Remove(args[0])
	},
}

func init() {
	aiMgmtCmd.AddCommand(mgmtGetCmd, mgmtListCmd, mgmtUpdateCmd, mgmtRemoveCmd)
}

// AIMgmtCmd is exported for registration in index.go.
var AIMgmtCmd = aiMgmtCmd
