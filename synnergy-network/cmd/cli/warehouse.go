package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"synnergy-network/core"
)

var (
	whOnce   bool
	wh       *core.Warehouse
	whLedger *core.Ledger
)

func whInit(cmd *cobra.Command, _ []string) error {
	if whOnce {
		return nil
	}
	_ = godotenv.Load()
	path := os.Getenv("LEDGER_PATH")
	if path == "" {
		path = "./ledger.db"
	}
	var err error
	whLedger, err = core.OpenLedger(path)
	if err != nil {
		return fmt.Errorf("open ledger: %w", err)
	}
	wh = core.NewWarehouse(whLedger)
	whOnce = true
	return nil
}

//---------------------------------------------------------------------
// Controllers
//---------------------------------------------------------------------

type warehouseController struct{}

func (warehouseController) Add(id, name string, qty uint64) error {
	ctx := &core.Context{Caller: core.Address{}}
	return wh.AddItem(ctx, id, name, qty)
}
func (warehouseController) Remove(id string) error {
	ctx := &core.Context{Caller: core.Address{}}
	return wh.RemoveItem(ctx, id)
}
func (warehouseController) Move(id, owner string) error {
	addr, err := core.ParseAddress(owner)
	if err != nil {
		return err
	}
	ctx := &core.Context{Caller: core.Address{}}
	return wh.MoveItem(ctx, id, addr)
}
func (warehouseController) List() error {
	items, err := wh.ListItems()
	if err != nil {
		return err
	}
	b, _ := json.MarshalIndent(items, "", "  ")
	fmt.Println(string(b))
	return nil
}

//---------------------------------------------------------------------
// CLI commands
//---------------------------------------------------------------------

var warehouseCmd = &cobra.Command{
	Use:               "warehouse",
	Short:             "Manage on-chain warehouse inventory",
	PersistentPreRunE: whInit,
}

var warehouseAddCmd = &cobra.Command{
	Use:   "add <id> <name> <qty>",
	Short: "Add a new item",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		qty, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return err
		}
		return warehouseController{}.Add(args[0], args[1], qty)
	},
}

var warehouseRemoveCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove an item",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return warehouseController{}.Remove(args[0])
	},
}

var warehouseMoveCmd = &cobra.Command{
	Use:   "move <id> <owner>",
	Short: "Change item owner",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return warehouseController{}.Move(args[0], args[1])
	},
}

var warehouseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all items",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return warehouseController{}.List()
	},
}

func init() {
	warehouseCmd.AddCommand(warehouseAddCmd)
	warehouseCmd.AddCommand(warehouseRemoveCmd)
	warehouseCmd.AddCommand(warehouseMoveCmd)
	warehouseCmd.AddCommand(warehouseListCmd)
}

// WarehouseCmd is the entry point for root.RegisterRoutes
var WarehouseCmd = warehouseCmd
