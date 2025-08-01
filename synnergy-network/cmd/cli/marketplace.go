package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"synnergy-network/core"
)

func mpParseAddr(hexStr string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(hexStr, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("invalid address")
	}
	copy(a[:], b)
	return a, nil
}

var marketCmd = &cobra.Command{
	Use:   "marketplace",
	Short: "General marketplace operations",
}

var mpListCreateCmd = &cobra.Command{
	Use:   "listing:create [price] [metadata-json]",
	Short: "Create a marketplace listing",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		price, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil || price == 0 {
			return fmt.Errorf("invalid price")
		}
		var meta map[string]string
		if err := json.Unmarshal([]byte(args[1]), &meta); err != nil {
			return fmt.Errorf("invalid meta JSON: %w", err)
		}
		listing := &core.MarketListing{Seller: core.ModuleAddress("cli"), Price: price, Meta: meta}
		if err := core.CreateMarketListing(listing); err != nil {
			return err
		}
		out, _ := json.MarshalIndent(listing, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(out))
		return nil
	},
}

var mpListGetCmd = &cobra.Command{
	Use:   "listing:get [id]",
	Short: "Get a marketplace listing",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		l, err := core.GetMarketListing(args[0])
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(l, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(out))
		return nil
	},
}

var mpListCmd = &cobra.Command{
	Use:   "listing:list",
	Short: "List marketplace listings",
	RunE: func(cmd *cobra.Command, args []string) error {
		lst, err := core.ListMarketListings(nil)
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(lst, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(out))
		return nil
	},
}

var mpBuyCmd = &cobra.Command{
	Use:   "buy [listing-id] [buyer]",
	Short: "Purchase a listing",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := mpParseAddr(args[1])
		if err != nil {
			return err
		}
		deal, err := core.PurchaseItem(&core.Context{}, args[0], addr)
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(deal, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(out))
		return nil
	},
}

var mpCancelCmd = &cobra.Command{
	Use:   "cancel [listing-id]",
	Short: "Cancel an open listing",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return core.CancelListing(args[0])
	},
}

var mpReleaseCmd = &cobra.Command{
	Use:   "release [escrow-id]",
	Short: "Release escrow funds",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return core.ReleaseFunds(&core.Context{}, args[0])
	},
}

var mpDealGetCmd = &cobra.Command{
	Use:   "deal:get [id]",
	Short: "Get deal details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		d, err := core.GetMarketDeal(args[0])
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(d, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(out))
		return nil
	},
}

var mpDealListCmd = &cobra.Command{
	Use:   "deal:list",
	Short: "List marketplace deals",
	RunE: func(cmd *cobra.Command, args []string) error {
		ds, err := core.ListMarketDeals(nil)
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(ds, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(out))
		return nil
	},
}

func init() {
	marketCmd.AddCommand(mpListCreateCmd)
	marketCmd.AddCommand(mpListGetCmd)
	marketCmd.AddCommand(mpListCmd)
	marketCmd.AddCommand(mpBuyCmd)
	marketCmd.AddCommand(mpCancelCmd)
	marketCmd.AddCommand(mpReleaseCmd)
	marketCmd.AddCommand(mpDealGetCmd)
	marketCmd.AddCommand(mpDealListCmd)
}

// MarketplaceCmd is the exported command group.
var MarketplaceCmd = marketCmd
