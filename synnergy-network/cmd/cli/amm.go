// cmd/cli/amm.go – Cobra CLI glue for the core/amm module.
// -----------------------------------------------------------
// Structure of this file
//   • Middleware (dependency wiring / guard‑rails)
//   • Controller (thin orchestrator around core.* helpers)
//   • CLI Commands   – declared top‑to‑bottom for discoverability
//   • Consolidation  – all commands mounted under root "amm" and
//                      exported via AMMCmd for import into your main index.
//
// Usage once injected into main root:
//     $ synnergy amm swap    <tokenIn> <amtIn> <tokenOut> <minOut>
//     $ synnergy amm add     <poolID> <provider> <amtA> <amtB>
//     $ synnergy amm remove  <poolID> <provider> <lpTokens>
//     $ synnergy amm quote   <tokenIn> <amtIn> <tokenOut>
//     $ synnergy amm pairs
// -----------------------------------------------------------
package cli

import (
    "context"
    "encoding/json"
    "fmt"
    "strconv"
    "strings"
    "time"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "go.uber.org/zap"

    core "github.com/synnergy-network/core" // adjust to your go.mod module path
)

//---------------------------------------------------------------------
// Middleware – executed for every ~amm command
//---------------------------------------------------------------------

type initMiddleware func(cmd *cobra.Command, args []string) error

// ensureAMMInitialised makes sure the AMM manager is ready and the
// routing graph built. It depends on other modules having initialised
// the ledger and pool manager – we assert their presence.
func ensureAMMInitialised(cmd *cobra.Command, _ []string) error {
    if core.Manager() != nil { // pools & router already set up
        return nil
    }
    // If Manager() is nil, attempt to initialise. This hooks into whatever
    // bootstrap is exposed by your application. We assume an InitPools(path)
    // environment variable pointing at a JSON fixture to bootstrap pools in
    // local‑dev and offline scenarios.  This logic is harmless when pools
    // already live on‑chain.

    fixture := viper.GetString("AMM_POOLS_FIXTURE")
    if fixture == "" {
        return fmt.Errorf("AMM manager not initialised – ensure blockchain node is running or set AMM_POOLS_FIXTURE for CLI offline mode")
    }
    if err := core.InitPoolsFromFile(fixture); err != nil { // hypothetical util
        return fmt.Errorf("init pools: %w", err)
    }
    zap.L().Sugar().Infow("AMM pools bootstrapped from fixture", "file", fixture)
    return nil
}

//---------------------------------------------------------------------
// Controller – provides user‑oriented façade, not exposing internals
//---------------------------------------------------------------------

type AMMController struct{}

func (c *AMMController) SwapExactIn(trader core.Address, tokenIn core.TokenID, amtIn uint64, tokenOut core.TokenID, minOut uint64, maxHops int) (uint64, error) {
    return core.SwapExactIn(trader, tokenIn, amtIn, tokenOut, minOut, maxHops)
}

func (c *AMMController) AddLiquidity(pid core.PoolID, provider core.Address, amtA, amtB uint64) (uint64, error) {
    return core.AddLiquidity(pid, provider, amtA, amtB)
}

func (c *AMMController) RemoveLiquidity(pid core.PoolID, provider core.Address, lp uint64) (uint64, uint64, error) {
    return core.RemoveLiquidity(pid, provider, lp)
}

func (c *AMMController) Quote(tokenIn core.TokenID, amtIn uint64, tokenOut core.TokenID, maxHops int) (uint64, error) {
    return core.Quote(tokenIn, amtIn, tokenOut, maxHops)
}

func (c *AMMController) AllPairs() [][2]core.TokenID { return core.AllPairs() }

//---------------------------------------------------------------------
// CLI command declarations
//---------------------------------------------------------------------

var ammCmd = &cobra.Command{
    Use:               "amm",
    Short:             "Automated‑market‑maker utilities (swap, liquidity, quotes)",
    PersistentPreRunE: ensureAMMInitialised,
}

// swap -----------------------------------------------------------------------
var swapCmd = &cobra.Command{
    Use:   "swap <tokenIn> <amtIn> <tokenOut> <minOut> [trader‑addr]",
    Short: "Swap exact tokens in for at least minOut received (best route)",
    Args:  cobra.RangeArgs(4, 5),
    RunE: func(cmd *cobra.Command, args []string) error {
        ctrl := &AMMController{}
        // mandatory
        tokenIn, tokenOut := core.TokenID(args[0]), core.TokenID(args[2])
        amtIn, err := strconv.ParseUint(args[1], 10, 64)
        if err != nil { return fmt.Errorf("amtIn uint64: %w", err) }
        minOut, err := strconv.ParseUint(args[3], 10, 64)
        if err != nil { return fmt.Errorf("minOut uint64: %w", err) }
        // optional trader address
        var trader core.Address = core.ModuleAddress("cli_trader")
        if len(args) == 5 { trader = core.Address(args[4]) }
        maxHops, _ := cmd.Flags().GetInt("max‑hops")
        out, err := ctrl.SwapExactIn(trader, tokenIn, amtIn, tokenOut, minOut, maxHops)
        if err != nil { return err }
        fmt.Printf("Received %d of %s\n", out, tokenOut)
        return nil
    },
}

// add‑liquidity ---------------------------------------------------------------
var addCmd = &cobra.Command{
    Use:   "add <poolID> <provider‑addr> <amtA> <amtB>",
    Short: "Add liquidity to an existing pool",
    Args:  cobra.ExactArgs(4),
    RunE: func(cmd *cobra.Command, args []string) error {
        ctrl := &AMMController{}
        pid := core.PoolID(args[0])
        provider := core.Address(args[1])
        amtA, err := strconv.ParseUint(args[2], 10, 64)
        if err != nil { return fmt.Errorf("amtA uint64: %w", err) }
        amtB, err := strconv.ParseUint(args[3], 10, 64)
        if err != nil { return fmt.Errorf("amtB uint64: %w", err) }
        lp, err := ctrl.AddLiquidity(pid, provider, amtA, amtB)
        if err != nil { return err }
        fmt.Printf("Minted %d LP tokens\n", lp)
        return nil
    },
}

// remove‑liquidity ------------------------------------------------------------
var removeCmd = &cobra.Command{
    Use:   "remove <poolID> <provider‑addr> <lpTokens>",
    Short: "Remove liquidity from a pool",
    Args:  cobra.ExactArgs(3),
    RunE: func(cmd *cobra.Command, args []string) error {
        ctrl := &AMMController{}
        pid := core.PoolID(args[0])
        provider := core.Address(args[1])
        lp, err := strconv.ParseUint(args[2], 10, 64)
        if err != nil { return fmt.Errorf("lpTokens uint64: %w", err) }
        outA, outB, err := ctrl.RemoveLiquidity(pid, provider, lp)
        if err != nil { return err }
        fmt.Printf("Redeemed %d / %d tokens from pool %s\n", outA, outB, pid)
        return nil
    },
}

// quote ----------------------------------------------------------------------
var quoteCmd = &cobra.Command{
    Use:   "quote <tokenIn> <amtIn> <tokenOut>",
    Short: "Estimate output amount without executing the swap",
    Args:  cobra.ExactArgs(3),
    RunE: func(cmd *cobra.Command, args []string) error {
        ctrl := &AMMController{}
        tokenIn, tokenOut := core.TokenID(args[0]), core.TokenID(args[2])
        amtIn, err := strconv.ParseUint(args[1], 10, 64)
        if err != nil { return fmt.Errorf("amtIn uint64: %w", err) }
        maxHops, _ := cmd.Flags().GetInt("max‑hops")
        out, err := ctrl.Quote(tokenIn, amtIn, tokenOut, maxHops)
        if err != nil { return err }
        fmt.Printf("Estimated output: %d of %s\n", out, tokenOut)
        return nil
    },
}

// pairs ----------------------------------------------------------------------
var pairsCmd = &cobra.Command{
    Use:   "pairs",
    Short: "List all tradable token pairs discovered by the router",
    Args:  cobra.NoArgs,
    RunE: func(cmd *cobra.Command, _ []string) error {
        ctrl := &AMMController{}
        pairs := ctrl.AllPairs()
        enc, _ := json.MarshalIndent(pairs, "", "  ")
        fmt.Println(string(enc))
        return nil
    },
}

//---------------------------------------------------------------------
// Consolidation & export
//---------------------------------------------------------------------

func init() {
    // shared flag across swap & quote
    swapCmd.Flags().Int("max‑hops", 4, "maximum hops allowed in the route")
    quoteCmd.Flags().Int("max‑hops", 4, "maximum hops allowed in the route")

    ammCmd.AddCommand(swapCmd)
    ammCmd.AddCommand(addCmd)
    ammCmd.AddCommand(removeCmd)
    ammCmd.AddCommand(quoteCmd)
    ammCmd.AddCommand(pairsCmd)
}

// Export for main‑index import: rootCmd.AddCommand(cli.AMMCmd)
var AMMCmd = ammCmd
