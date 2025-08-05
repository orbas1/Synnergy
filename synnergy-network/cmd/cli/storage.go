package cli

// cmd/cli/storage.go — CLI wrapper for the core/storage subsystem.
// ----------------------------------------------------------------------------
// Layout
//   1. Globals & middleware (env‑driven wiring of logger, ledger, storage).
//   2. Controllers – one per CLI sub‑command, thin and validated.
//   3. CLI definitions – commands + flags (TOP of file for discoverability).
//   4. Consolidated route export (BOTTOM), ready for import in root CLI.
// ----------------------------------------------------------------------------

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	logrus "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"synnergy-network/core"
)

// ---------------------------------------------------------------------------
// Globals & middleware
// ---------------------------------------------------------------------------

var (
	storage      *core.Storage
	storageLG    = logrus.New()
	storageFlags struct {
		ledgerPath   string
		gatewayURL   string
		cacheDir     string
		cacheEntries int
		timeoutSec   int
	}
)

func initStorageMiddleware(cmd *cobra.Command, args []string) {
	// 1) .env overrides
	_ = godotenv.Load()

	resolveStringFlag(cmd, "ledger", &storageFlags.ledgerPath, os.Getenv("LEDGER_PATH"))
	resolveStringFlag(cmd, "gateway", &storageFlags.gatewayURL, os.Getenv("IPFS_GATEWAY"))
	resolveStringFlag(cmd, "cache", &storageFlags.cacheDir, os.Getenv("CACHE_DIR"))
	resolveIntFlag(cmd, "cacheEntries", &storageFlags.cacheEntries, envInt("CACHE_ENTRIES", 10_000))
	resolveIntFlag(cmd, "timeout", &storageFlags.timeoutSec, envInt("GATEWAY_TIMEOUT", 30))

	if storageFlags.gatewayURL == "" {
		log.Fatalf("IPFS gateway URL must be provided via --gateway or IPFS_GATEWAY")
	}

	// 2) ledger setup (optional, but recommended)
	if storageFlags.ledgerPath == "" {
		exe, _ := os.Executable()
		storageFlags.ledgerPath = filepath.Join(filepath.Dir(exe), "ledger")
	}
	led, err := core.OpenLedger(storageFlags.ledgerPath)
	if err != nil {
		log.Fatalf("ledger open: %v", err)
	}

	// 3) build storage config
	cfg := &core.StorageConfig{
		IPFSGateway:      storageFlags.gatewayURL,
		CacheDir:         storageFlags.cacheDir,
		CacheSizeEntries: storageFlags.cacheEntries,
		GatewayTimeout:   time.Duration(storageFlags.timeoutSec) * time.Second,
	}

	storage, err = core.NewStorage(cfg, storageLG, led)
	if err != nil {
		log.Fatalf("storage init: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Controller helpers
// ---------------------------------------------------------------------------

func parseStorageAddress(hexStr string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(hexStr)
	if err != nil || len(b) != len(a) {
		return a, errors.New("address must be 20-byte hex")
	}
	copy(a[:], b)
	return a, nil
}

func storageBail(err error) {
	if err != nil {
		log.Fatalf("❌ %v", err)
	}
}

// ---------------------------------------------------------------------------
// Controllers – Pin & Retrieve
// ---------------------------------------------------------------------------

func pinHandler(cmd *cobra.Command, args []string) {
	file, _ := cmd.Flags().GetString("file")
	payerHex, _ := cmd.Flags().GetString("payer")

	if file == "" {
		_ = cmd.Usage()
		storageBail(errors.New("--file is required"))
	}
	if payerHex == "" {
		_ = cmd.Usage()
		storageBail(errors.New("--payer is required"))
	}

	data, err := os.ReadFile(file)
	storageBail(err)
	payer, err := parseStorageAddress(payerHex)
	storageBail(err)

	ctx, cancel := context.WithTimeout(cmd.Context(), time.Duration(storageFlags.timeoutSec)*time.Second)
	defer cancel()

	cid, size, err := storage.Pin(ctx, data, payer)
	storageBail(err)
	fmt.Printf("✅ pinned %s (%.2f KB)\n", cid, float64(size)/1024)
}

func getHandler(cmd *cobra.Command, args []string) {
	cidStr, _ := cmd.Flags().GetString("cid")
	outPath, _ := cmd.Flags().GetString("out")

	if cidStr == "" {
		_ = cmd.Usage()
		storageBail(errors.New("--cid is required"))
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), time.Duration(storageFlags.timeoutSec)*time.Second)
	defer cancel()

	data, err := storage.Retrieve(ctx, cidStr)
	storageBail(err)

	if outPath == "-" || outPath == "" {
		_, _ = io.Copy(os.Stdout, bytes.NewReader(data))
		return
	}
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		storageBail(err)
	}
	fmt.Printf("✅ wrote %d bytes to %s\n", len(data), outPath)
}

// ---------------------------------------------------------------------------
// Controllers – Listings & Deals
// ---------------------------------------------------------------------------

func createListingHandler(cmd *cobra.Command, args []string) {
	providerHex, _ := cmd.Flags().GetString("provider")
	priceStr, _ := cmd.Flags().GetString("price")
	capacity, _ := cmd.Flags().GetInt("capacity")

	if providerHex == "" || priceStr == "" || capacity == 0 {
		_ = cmd.Usage()
		storageBail(errors.New("--provider, --price and --capacity are required"))
	}

	provider, err := parseStorageAddress(providerHex)
	storageBail(err)
	price, err := strconv.ParseUint(priceStr, 10, 64)
	storageBail(err)

	listing := &core.StorageListing{
		Provider:   provider,
		PricePerGB: price,
		CapacityGB: capacity,
	}
	storageBail(core.CreateListing(listing))
	fmt.Printf("✅ listing created: %s\n", listing.ID)
}

func openDealHandler(cmd *cobra.Command, args []string) {
	listingID, _ := cmd.Flags().GetString("listing")
	clientHex, _ := cmd.Flags().GetString("client")
	durHours, _ := cmd.Flags().GetInt("duration")

	if listingID == "" || clientHex == "" || durHours == 0 {
		_ = cmd.Usage()
		storageBail(errors.New("--listing, --client and --duration are required"))
	}

	client, err := parseStorageAddress(clientHex)
	storageBail(err)

	deal := &core.StorageDeal{
		ListingID: listingID,
		Client:    client,
		Duration:  time.Duration(durHours) * time.Hour,
	}
	esc, err := core.OpenDeal(deal)
	storageBail(err)
	fmt.Printf("✅ deal opened: %s  escrow=%s\n", deal.ID, esc.ID)
}

func closeDealHandler(cmd *cobra.Command, args []string) {
	dealID, _ := cmd.Flags().GetString("deal")
	if dealID == "" {
		_ = cmd.Usage()
		storageBail(errors.New("--deal is required"))
	}
	ctx := &core.Context{} // assuming a Core tx context implementation
	storageBail(core.CloseDeal(ctx, dealID))
	fmt.Println("✅ deal closed")
}

func getListingHandler(cmd *cobra.Command, args []string) {
	id, _ := cmd.Flags().GetString("id")
	if id == "" {
		_ = cmd.Usage()
		storageBail(errors.New("--id is required"))
	}
	listing, err := core.GetListing(id)
	storageBail(err)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(listing)
}

func listListingsHandler(cmd *cobra.Command, args []string) {
	provHex, _ := cmd.Flags().GetString("provider")
	var prov *core.Address
	if provHex != "" {
		a, err := parseStorageAddress(provHex)
		storageBail(err)
		prov = &a
	}
	listings, err := core.ListListings(prov)
	storageBail(err)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(listings)
}

func getDealHandler(cmd *cobra.Command, args []string) {
	id, _ := cmd.Flags().GetString("id")
	if id == "" {
		_ = cmd.Usage()
		storageBail(errors.New("--id is required"))
	}
	deal, err := core.GetDeal(id)
	storageBail(err)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(deal)
}

func listDealsHandler(cmd *cobra.Command, args []string) {
	provHex, _ := cmd.Flags().GetString("provider")
	clientHex, _ := cmd.Flags().GetString("client")
	var prov *core.Address
	var client *core.Address
	if provHex != "" {
		a, err := parseStorageAddress(provHex)
		storageBail(err)
		prov = &a
	}
	if clientHex != "" {
		a, err := parseStorageAddress(clientHex)
		storageBail(err)
		client = &a
	}
	deals, err := core.ListDeals(prov, client)
	storageBail(err)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(deals)
}

// ---------------------------------------------------------------------------
// CLI definitions (TOP section)
// ---------------------------------------------------------------------------

var storageCmd = &cobra.Command{
	Use:              "storage",
	Short:            "Decentralised storage operations (IPFS/Arweave & marketplace)",
	PersistentPreRun: initStorageMiddleware,
}

var pinCmd = &cobra.Command{
	Use:   "pin",
	Short: "Pin file/data to the configured gateway",
	Run:   pinHandler,
}

var storageGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve data by CID (cache → gateway)",
	Run:   getHandler,
}

var listCreateCmd = &cobra.Command{
	Use:   "listing:create",
	Short: "Create a storage listing (provider side)",
	Run:   createListingHandler,
}

var listGetCmd = &cobra.Command{
	Use:   "listing:get",
	Short: "Get a storage listing by ID",
	Run:   getListingHandler,
}

var listListCmd = &cobra.Command{
	Use:   "listing:list",
	Short: "List storage listings",
	Run:   listListingsHandler,
}

var dealOpenCmd = &cobra.Command{
	Use:   "deal:open",
	Short: "Open a storage deal backed by escrow (client side)",
	Run:   openDealHandler,
}

var dealCloseCmd = &cobra.Command{
	Use:   "deal:close",
	Short: "Close a storage deal and release escrow (provider side)",
	Run:   closeDealHandler,
}

var dealGetCmd = &cobra.Command{
	Use:   "deal:get",
	Short: "Get storage deal details",
	Run:   getDealHandler,
}

var dealListCmd = &cobra.Command{
	Use:   "deal:list",
	Short: "List storage deals",
	Run:   listDealsHandler,
}

func init() {
	// persistent root flags (shared)
	storageCmd.PersistentFlags().String("ledger", "", "Path to ledger DB (LEDGER_PATH)")
	storageCmd.PersistentFlags().String("gateway", "", "IPFS gateway base URL (IPFS_GATEWAY)")
	storageCmd.PersistentFlags().String("cache", os.TempDir(), "Cache directory (CACHE_DIR)")
	storageCmd.PersistentFlags().Int("cacheEntries", 10_000, "Max cache entries (CACHE_ENTRIES)")
	storageCmd.PersistentFlags().Int("timeout", 30, "Gateway timeout seconds (GATEWAY_TIMEOUT)")

	// pin flags
	pinCmd.Flags().String("file", "", "Path to file to pin [required]")
	pinCmd.Flags().String("payer", "", "Address paying storage rent (hex) [required]")

	// get flags
	storageGetCmd.Flags().String("cid", "", "Content identifier to fetch [required]")
	storageGetCmd.Flags().String("out", "-", "Output file or '-' for STDOUT")

	// listing flags
	listCreateCmd.Flags().String("provider", "", "Provider address (hex) [required]")
	listCreateCmd.Flags().String("price", "", "Price per GB in tokens [required]")
	listCreateCmd.Flags().Int("capacity", 0, "Capacity in GB [required]")
	listGetCmd.Flags().String("id", "", "Listing ID [required]")
	listListCmd.Flags().String("provider", "", "Filter by provider (hex)")

	// deal open flags
	dealOpenCmd.Flags().String("listing", "", "Listing ID [required]")
	dealOpenCmd.Flags().String("client", "", "Client address (hex) [required]")
	dealOpenCmd.Flags().Int("duration", 0, "Deal duration hours [required]")

	// deal close flags
	dealCloseCmd.Flags().String("deal", "", "Deal ID [required]")
	dealGetCmd.Flags().String("id", "", "Deal ID [required]")
	dealListCmd.Flags().String("provider", "", "Filter by provider (hex)")
	dealListCmd.Flags().String("client", "", "Filter by client (hex)")

	// register sub‑commands
	storageCmd.AddCommand(pinCmd)
	storageCmd.AddCommand(storageGetCmd)
	storageCmd.AddCommand(listCreateCmd)
	storageCmd.AddCommand(listGetCmd)
	storageCmd.AddCommand(listListCmd)
	storageCmd.AddCommand(dealOpenCmd)
	storageCmd.AddCommand(dealCloseCmd)
	storageCmd.AddCommand(dealGetCmd)
	storageCmd.AddCommand(dealListCmd)
}

// ---------------------------------------------------------------------------
// Helpers – env handling
// ---------------------------------------------------------------------------

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func resolveStringFlag(cmd *cobra.Command, name string, target *string, fallback string) {
	if v, _ := cmd.Flags().GetString(name); v != "" {
		*target = v
	} else if fallback != "" {
		*target = fallback
	}
}

func resolveIntFlag(cmd *cobra.Command, name string, target *int, fallback int) {
	if v, _ := cmd.Flags().GetInt(name); v != 0 {
		*target = v
	} else {
		*target = fallback
	}
}

// ---------------------------------------------------------------------------
// Consolidated route export (BOTTOM) — importable by root CLI.
// ---------------------------------------------------------------------------

// StorageRoute represents the entry‑point command (root: "storage").
var StorageRoute = storageCmd

// ---------------------------------------------------------------------------
// END cmd/cli/storage.go
// ---------------------------------------------------------------------------
