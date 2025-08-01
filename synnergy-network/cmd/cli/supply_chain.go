package cli

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

// Middleware ensures the KV store is initialised.
func supplyMiddleware(cmd *cobra.Command, args []string) error {
	if core.CurrentStore() == nil {
		return errors.New("supply chain store not initialised")
	}
	return nil
}

// Controller wraps core supply chain helpers.
type SupplyController struct{}

func (c *SupplyController) Register(id, desc, ownerHex, loc string) error {
	ownerBytes, err := hex.DecodeString(ownerHex)
	if err != nil || len(ownerBytes) != 20 {
		return fmt.Errorf("invalid owner address")
	}
	var addr core.Address
	copy(addr[:], ownerBytes)
	item := core.SupplyItem{ID: id, Description: desc, Owner: addr, Location: loc}
	return core.RegisterItem(item)
}

func (c *SupplyController) UpdateLocation(id, loc string) error {
	return core.UpdateLocation(id, loc)
}

func (c *SupplyController) MarkStatus(id, status string) error {
	return core.MarkStatus(id, status)
}

func (c *SupplyController) Get(id string) (*core.SupplyItem, error) {
	return core.GetItem(id)
}

// CLI commands
var (
	supplyCmd = &cobra.Command{
		Use:               "supply",
		Short:             "Manage on-chain supply chain records",
		PersistentPreRunE: supplyMiddleware,
	}

	supplyRegisterCmd = &cobra.Command{
		Use:   "register <id> <desc> <owner> <location>",
		Short: "Register a new supply item",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctrl := &SupplyController{}
			return ctrl.Register(args[0], args[1], args[2], args[3])
		},
	}

	supplyUpdateCmd = &cobra.Command{
		Use:   "update-location <id> <location>",
		Short: "Update item location",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctrl := &SupplyController{}
			return ctrl.UpdateLocation(args[0], args[1])
		},
	}

	supplyStatusCmd = &cobra.Command{
		Use:   "status <id> <status>",
		Short: "Update item status",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctrl := &SupplyController{}
			return ctrl.MarkStatus(args[0], args[1])
		},
	}

	supplyGetCmd = &cobra.Command{
		Use:   "get <id>",
		Short: "Get item details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctrl := &SupplyController{}
			item, err := ctrl.Get(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%+v\n", *item)
			return nil
		},
	}
)

func init() {
	supplyCmd.AddCommand(supplyRegisterCmd)
	supplyCmd.AddCommand(supplyUpdateCmd)
	supplyCmd.AddCommand(supplyStatusCmd)
	supplyCmd.AddCommand(supplyGetCmd)
}

var SupplyCmd = supplyCmd
