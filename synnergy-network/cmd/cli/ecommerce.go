package cli

// Basic ecommerce CLI commands allowing the creation of marketplace listings and
// the purchase of items. These commands are intentionally lightweight and use
// the core Ecommerce module which stores state inside the ledger.

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var (
	ecOnce sync.Once
	ec     *core.Ecommerce
	ecLed  *core.Ledger
)

func ecInit(cmd *cobra.Command, _ []string) error {
	var err error
	ecOnce.Do(func() {
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		ecLed, err = core.OpenLedger(path)
		if err != nil {
			return
		}
		ec = core.NewEcommerce(ecLed)
	})
	return err
}

// ecommerceCmd is the root command for marketplace operations.
var EcommerceCmd = &cobra.Command{
	Use:               "ecommerce",
	Short:             "Simple marketplace utilities",
	PersistentPreRunE: ecInit,
}

var ecListCmd = &cobra.Command{
	Use:   "list [token] [price] [qty] [seller]",
	Short: "Create a new listing",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		token := args[0]
		price, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return err
		}
		qty, err := strconv.ParseUint(args[2], 10, 32)
		if err != nil {
			return err
		}
		var seller core.Address
		b, err := hex.DecodeString(strings.TrimPrefix(args[3], "0x"))
		if err != nil || len(b) != len(seller) {
			return fmt.Errorf("invalid seller address")
		}
		copy(seller[:], b)
		id, err := ec.CreateListing(seller, token, price, uint32(qty))
		if err != nil {
			return err
		}
		fmt.Printf("listing created with id %d\n", id)
		return nil
	},
}

var ecBuyCmd = &cobra.Command{
	Use:   "buy [id] [qty] [buyer]",
	Short: "Purchase an item",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return err
		}
		qty, err := strconv.ParseUint(args[1], 10, 32)
		if err != nil {
			return err
		}
		var buyer core.Address
		b, err := hex.DecodeString(strings.TrimPrefix(args[2], "0x"))
		if err != nil || len(b) != len(buyer) {
			return fmt.Errorf("invalid buyer address")
		}
		copy(buyer[:], b)
		return ec.PurchaseItem(buyer, id, uint32(qty))
	},
}

var ecViewCmd = &cobra.Command{
	Use:   "view [id]",
	Short: "View a listing",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return err
		}
		lst, err := ec.GetListing(id)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(lst)
	},
}

func init() {
	EcommerceCmd.AddCommand(ecListCmd, ecBuyCmd, ecViewCmd)
}

// END ecommerce.go
