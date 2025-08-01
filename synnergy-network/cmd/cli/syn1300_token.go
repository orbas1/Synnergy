package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

// SupplyChainTokenController provides convenience wrappers around the core token.
type SupplyChainTokenController struct {
	token *core.SupplyChainToken
}

func (c *SupplyChainTokenController) Register(id, desc, loc string, owner core.Address) error {
	asset := core.SupplyChainAsset{ID: id, Description: desc, Location: loc, Status: "created", Owner: owner}
	return c.token.RegisterAsset(asset)
}

func (c *SupplyChainTokenController) UpdateLocation(id, loc string) error {
	return c.token.UpdateLocation(id, loc)
}

func (c *SupplyChainTokenController) UpdateStatus(id, status string) error {
	return c.token.UpdateStatus(id, status)
}

func (c *SupplyChainTokenController) Transfer(id string, to core.Address) error {
	return c.token.TransferAsset(id, to)
}

// CLI command definitions
var (
	scTokenCmd = &cobra.Command{
		Use:   "sctoken",
		Short: "Manage SYN1300 supply chain tokens",
	}
)

// placeholder command examples
var scTokenInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "show token info",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "SYN1300 supply chain token")
		return nil
	},
}

func init() {
	scTokenCmd.AddCommand(scTokenInfoCmd)
	TokensCmd.AddCommand(scTokenCmd)
}
