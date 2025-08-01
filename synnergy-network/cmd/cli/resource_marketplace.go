package cli

// resource_marketplace.go -- CLI bindings for the compute resource marketplace
// module. Providers can list hardware resources and clients open rental deals
// backed by escrow in the ledger.

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"synnergy-network/core"
)

// helpers -------------------------------------------------------------------
func parseResAddr(hexStr string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(hexStr)
	if err != nil || len(b) != len(a) {
		return a, errors.New("address must be 20-byte hex")
	}
	copy(a[:], b)
	return a, nil
}

func resBail(err error) {
	if err != nil {
		panic(fmt.Errorf("❌ %v", err))
	}
}

// controllers ---------------------------------------------------------------
func resListCreate(cmd *cobra.Command, args []string) {
	providerHex, _ := cmd.Flags().GetString("provider")
	priceStr, _ := cmd.Flags().GetString("price")
	units, _ := cmd.Flags().GetInt("units")
	if providerHex == "" || priceStr == "" || units == 0 {
		_ = cmd.Usage()
		resBail(errors.New("--provider, --price and --units are required"))
	}
	provider, err := parseResAddr(providerHex)
	resBail(err)
	price, err := strconv.ParseUint(priceStr, 10, 64)
	resBail(err)

	l := &core.ResourceListing{Provider: provider, PricePerHour: price, Units: units}
	resBail(core.ListResource(l))
	fmt.Printf("✅ resource listing created: %s\n", l.ID)
}

func resDealOpen(cmd *cobra.Command, args []string) {
	listingID, _ := cmd.Flags().GetString("listing")
	clientHex, _ := cmd.Flags().GetString("client")
	durH, _ := cmd.Flags().GetInt("duration")
	if listingID == "" || clientHex == "" || durH == 0 {
		_ = cmd.Usage()
		resBail(errors.New("--listing, --client and --duration required"))
	}
	client, err := parseResAddr(clientHex)
	resBail(err)
	d := &core.ResourceDeal{ListingID: listingID, Client: client, Duration: time.Duration(durH) * time.Hour}
	esc, err := core.OpenResourceDeal(d)
	resBail(err)
	fmt.Printf("✅ deal opened: %s escrow=%s\n", d.ID, esc.ID)
}

func resDealClose(cmd *cobra.Command, args []string) {
	dealID, _ := cmd.Flags().GetString("deal")
	if dealID == "" {
		_ = cmd.Usage()
		resBail(errors.New("--deal required"))
	}
	resBail(core.CloseResourceDeal(&core.Context{}, dealID))
	fmt.Println("✅ deal closed")
}

func resListGet(cmd *cobra.Command, args []string) {
	id, _ := cmd.Flags().GetString("id")
	if id == "" {
		_ = cmd.Usage()
		resBail(errors.New("--id required"))
	}
	l, err := core.GetResourceListing(id)
	resBail(err)
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	_ = enc.Encode(l)
}

func resListList(cmd *cobra.Command, args []string) {
	provHex, _ := cmd.Flags().GetString("provider")
	var p *core.Address
	if provHex != "" {
		a, err := parseResAddr(provHex)
		resBail(err)
		p = &a
	}
	ls, err := core.ListResourceListings(p)
	resBail(err)
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	_ = enc.Encode(ls)
}

func resDealGet(cmd *cobra.Command, args []string) {
	id, _ := cmd.Flags().GetString("id")
	if id == "" {
		_ = cmd.Usage()
		resBail(errors.New("--id required"))
	}
	d, err := core.GetResourceDeal(id)
	resBail(err)
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	_ = enc.Encode(d)
}

func resDealList(cmd *cobra.Command, args []string) {
	provHex, _ := cmd.Flags().GetString("provider")
	cliHex, _ := cmd.Flags().GetString("client")
	var p *core.Address
	var c *core.Address
	if provHex != "" {
		a, err := parseResAddr(provHex)
		resBail(err)
		p = &a
	}
	if cliHex != "" {
		a, err := parseResAddr(cliHex)
		resBail(err)
		c = &a
	}
	ds, err := core.ListResourceDeals(p, c)
	resBail(err)
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	_ = enc.Encode(ds)
}

// command setup ------------------------------------------------------------
var resCmd = &cobra.Command{Use: "resource", Short: "Resource marketplace operations"}
var resListCreateCmd = &cobra.Command{Use: "listing:create", Run: resListCreate}
var resListGetCmd = &cobra.Command{Use: "listing:get", Run: resListGet}
var resListListCmd = &cobra.Command{Use: "listing:list", Run: resListList}
var resDealOpenCmd = &cobra.Command{Use: "deal:open", Run: resDealOpen}
var resDealCloseCmd = &cobra.Command{Use: "deal:close", Run: resDealClose}
var resDealGetCmd = &cobra.Command{Use: "deal:get", Run: resDealGet}
var resDealListCmd = &cobra.Command{Use: "deal:list", Run: resDealList}

func init() {
	resListCreateCmd.Flags().String("provider", "", "provider address [hex]")
	resListCreateCmd.Flags().String("price", "", "price per hour")
	resListCreateCmd.Flags().Int("units", 0, "number of units")
	resListGetCmd.Flags().String("id", "", "listing id")
	resListListCmd.Flags().String("provider", "", "filter by provider")
	resDealOpenCmd.Flags().String("listing", "", "listing id")
	resDealOpenCmd.Flags().String("client", "", "client address")
	resDealOpenCmd.Flags().Int("duration", 0, "hours")
	resDealCloseCmd.Flags().String("deal", "", "deal id")
	resDealGetCmd.Flags().String("id", "", "deal id")
	resDealListCmd.Flags().String("provider", "", "filter provider")
	resDealListCmd.Flags().String("client", "", "filter client")

	resCmd.AddCommand(resListCreateCmd, resListGetCmd, resListListCmd,
		resDealOpenCmd, resDealCloseCmd, resDealGetCmd, resDealListCmd)
}

// ResourceCmd exported to index.go
var ResourceCmd = resCmd
