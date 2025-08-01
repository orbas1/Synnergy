// cmd/cli/ai.go – Cobra CLI glue for the core/ai module.
// -------------------------------------------------------------
// The file follows a layered structure:
//   • Middleware                   – dependency wiring & guard rails
//   • Controller                   – light service wrapper around core logic
//   • Route declarations           – one var per CLI command
//   • Consolidation & export       – all commands wired to the ~ai root and
//                                    exported via AICmd for main‑index import.
//
//   Example usage once mounted in main Index CLI:
//       $ synnergy ai predict ./tx.json
//       $ synnergy ai publish Qm… --royalty 50
//       $ synnergy ai list
// -------------------------------------------------------------

package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	core "synnergy-network/core" // adjust if your go.mod path differs
)

//---------------------------------------------------------------------
// Middleware – executed for every ~ai command
//---------------------------------------------------------------------

type middleware func(cmd *cobra.Command, args []string) error

// ensureAIInitialised wires the core.AI() singleton using env‑config.
func ensureAIInitialised(cmd *cobra.Command, _ []string) error {
	if core.AI() != nil {
		return nil // already ready
	}

	// bootstrap logger early – using zap, fall back on stdout if this fails.
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)

	grpcEndpoint := viper.GetString("AI_GRPC_ENDPOINT")
	if grpcEndpoint == "" {
		return fmt.Errorf("AI_GRPC_ENDPOINT not set in environment or config")
	}

	// ledger connection depends on wider app – obtain via container / DI.
	// The compiled gRPC stub lives in core; we inject a concrete client impl.
	client := core.NewTFStubClient(grpcEndpoint)
	if err := core.InitAI(nil, grpcEndpoint, client); err != nil {
		return fmt.Errorf("init AI engine: %w", err)
	}
	zap.L().Sugar().Infow("AI engine ready", "endpoint", grpcEndpoint)
	return nil
}

//---------------------------------------------------------------------
// Controller – wraps core API + common utilities
//---------------------------------------------------------------------

type AIController struct{}

func parseAddr(hexStr string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(hexStr, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("invalid address")
	}
	copy(a[:], b)
	return a, nil
}

func (c *AIController) PredictAnomaly(txPath string) (float32, error) {
	raw, err := ioutil.ReadFile(txPath)
	if err != nil {
		return 0, fmt.Errorf("read tx file: %w", err)
	}
	var tx core.Transaction
	if err := json.Unmarshal(raw, &tx); err != nil {
		return 0, fmt.Errorf("decode tx JSON: %w", err)
	}
	return core.AI().PredictAnomaly(&tx)
}

func (c *AIController) OptimiseFees(statsPath string) (uint64, error) {
	raw, err := ioutil.ReadFile(statsPath)
	if err != nil {
		return 0, fmt.Errorf("read stats file: %w", err)
	}
	var stats []core.BlockStats
	if err := json.Unmarshal(raw, &stats); err != nil {
		return 0, fmt.Errorf("decode stats JSON: %w", err)
	}
	return core.AI().OptimizeFees(stats)
}

func (c *AIController) PredictVolume(path string) (uint64, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read volume file: %w", err)
	}
	var vols []core.TxVolume
	if err := json.Unmarshal(raw, &vols); err != nil {
		return 0, fmt.Errorf("decode volume JSON: %w", err)
	}
	return core.AI().PredictVolume(vols)
}

func (c *AIController) PublishModel(cid string, royalty uint16) ([32]byte, error) {
	// For demo we assume creator == ModuleAddress("cli") – adapt to wallet
	creator := core.ModuleAddress("cli")
	return core.AI().PublishModel(cid, creator, royalty)
}

func (c *AIController) FetchModel(hashHex string) (core.ModelMeta, error) {
	hRaw, err := hex.DecodeString(strings.TrimPrefix(hashHex, "0x"))
	if err != nil || len(hRaw) != 32 {
		return core.ModelMeta{}, fmt.Errorf("invalid hash – want 32‑byte hex string")
	}
	var h [32]byte
	copy(h[:], hRaw)
	return core.AI().FetchModel(h)
}

// Marketplace helpers ------------------------------------------------
func (c *AIController) ListModel(price uint64, cid string) (*core.ModelListing, error) {
	listing := &core.ModelListing{
		Seller: core.ModuleAddress("cli"),
		Price:  price,
		Meta:   map[string]string{"cid": cid},
	}
	if err := core.ListModel(listing); err != nil {
		return nil, err
	}
	return listing, nil
}

//---------------------------------------------------------------------
// Route declarations – one per sub‑command
//---------------------------------------------------------------------

var aiCmd = &cobra.Command{
	Use:               "ai",
	Short:             "On‑chain AI utilities (fraud, fees, model marketplace)",
	PersistentPreRunE: ensureAIInitialised,
}

// predict --------------------------------------------------------------------
var aiPredictCmd = &cobra.Command{
	Use:   "predict [tx.json]",
	Short: "Predict fraud probability for a transaction",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AIController{}
		score, err := ctrl.PredictAnomaly(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Fraud risk score: %.4f\n", score)
		return nil
	},
}

// optimise -------------------------------------------------------------------
var optimiseCmd = &cobra.Command{
	Use:   "optimise [stats.json]",
	Short: "Suggest optimal basefee for upcoming block",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AIController{}
		target, err := ctrl.OptimiseFees(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Suggested basefee: %d wei\n", target)
		return nil
	},
}

// volume --------------------------------------------------------------------
var volumeCmd = &cobra.Command{
	Use:   "volume [stats.json]",
	Short: "Predict upcoming transaction volume",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AIController{}
		count, err := ctrl.PredictVolume(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Predicted volume: %d txs\n", count)
		return nil
	},
}

// publish --------------------------------------------------------------------
var publishCmd = &cobra.Command{
	Use:   "publish [cid]",
	Short: "Publish a new model hash with optional royalty (bp)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		royalty, _ := cmd.Flags().GetUint16("royalty")
		ctrl := &AIController{}
		h, err := ctrl.PublishModel(args[0], royalty)
		if err != nil {
			return err
		}
		fmt.Printf("Model published. Hash: %x\n", h[:])
		return nil
	},
}

// fetch ----------------------------------------------------------------------
var fetchCmd = &cobra.Command{
	Use:   "fetch [model‑hash]",
	Short: "Fetch model metadata by SHA‑256 hash",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AIController{}
		meta, err := ctrl.FetchModel(args[0])
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(meta, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

// list -----------------------------------------------------------------------
var aiListCmd = &cobra.Command{
	Use:   "list [price] [cid]",
	Short: "Create a marketplace listing for a model",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		price, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("price must be uint64: %w", err)
		}
		ctrl := &AIController{}
		listing, err := ctrl.ListModel(price, args[1])
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(listing, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

// buy ------------------------------------------------------------------------
var buyCmd = &cobra.Command{
	Use:   "buy [listing‑id] [buyer‑addr]",
	Short: "Buy a model – funds held in escrow",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := parseAddr(args[1])
		if err != nil {
			return err
		}
		esc, err := core.BuyModel(&core.Context{}, args[0], addr)
		if err != nil {
			return err
		}
		fmt.Printf("Escrow %s funded. Buyer %s\n", esc.ID, esc.Buyer)
		return nil
	},
}

// rent -----------------------------------------------------------------------
var rentCmd = &cobra.Command{
	Use:   "rent [listing‑id] [renter‑addr] [hours]",
	Short: "Rent a model for N hours – escrow held until end",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		hours, err := strconv.Atoi(args[2])
		if err != nil || hours <= 0 {
			return fmt.Errorf("hours must be positive int: %w", err)
		}
		addr, err := parseAddr(args[1])
		if err != nil {
			return err
		}
		esc, err := core.RentModel(&core.Context{}, args[0], addr, time.Duration(hours)*time.Hour)
		if err != nil {
			return err
		}
		fmt.Printf("Escrow %s funded for rental. Renter %s\n", esc.ID, esc.Buyer)
		return nil
	},
}

// release‑escrow --------------------------------------------------------------
var releaseCmd = &cobra.Command{
	Use:   "release [escrow‑id]",
	Short: "Release funds from escrow to seller (admin/op) ",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := core.ReleaseEscrow(&core.Context{}, args[0]); err != nil {
			return err
		}
		fmt.Printf("Escrow %s released.\n", args[0])
		return nil
	},
}

//---------------------------------------------------------------------
// Consolidation & export – sub‑routes wired to root, then exported
//---------------------------------------------------------------------

func init() {
	// bind flags
	publishCmd.Flags().Uint16("royalty", 0, "royalty basis points (max 1000 = 10%)")

	// attach sub‑routes
	aiCmd.AddCommand(aiPredictCmd)
	aiCmd.AddCommand(optimiseCmd)
	aiCmd.AddCommand(volumeCmd)
	aiCmd.AddCommand(publishCmd)
	aiCmd.AddCommand(fetchCmd)
	aiCmd.AddCommand(aiListCmd)
	aiCmd.AddCommand(buyCmd)
	aiCmd.AddCommand(rentCmd)
	aiCmd.AddCommand(releaseCmd)
}

// Exported for main index CLI: rootCmd.AddCommand(cli.AICmd)
var AICmd = aiCmd
