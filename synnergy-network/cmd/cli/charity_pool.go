// cmd/cli/charity_pool.go – Cobra CLI integration for the on‑chain CharityPool
// -----------------------------------------------------------------------------
// Layout
//   • Middleware        – dependency & genesis wiring
//   • Controller        – thin wrapper around core.CharityPool
//   • Command decl.     – user‑visible sub‑commands at top of file
//   • Consolidation     – all commands mounted under root `charity` and
//                         exported via CharityCmd for main‑index import.
//
// After mounting in your root command:
//     $ synnergy charity register tz1Charity wildlife "Save the Whales"
//     $ synnergy charity vote tz1Alice tz1Charity
//     $ synnergy charity tick                     # manual cron kick
//     $ synnergy charity winners                  # inspect winners list
// -----------------------------------------------------------------------------
package cli

import (
    "encoding/json"
    "errors"
    "fmt"
    "strconv"
    "strings"
    "time"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "github.com/sirupsen/logrus"
    core "synnergy-network/core" // adjust if go.mod path differs
)

//---------------------------------------------------------------------
// Constants duplicated locally (unexported in core)
//---------------------------------------------------------------------

const (
    cycleDuration      = 90 * 24 * time.Hour
    registrationCutoff = 30 * 24 * time.Hour
    votingWindow       = 15 * 24 * time.Hour
)

//---------------------------------------------------------------------
// Middleware – initialises a singleton CharityPool backed by the ledger
//---------------------------------------------------------------------

var (
    cp        *core.CharityPool
    genesisTs time.Time
)

type cliElectorate struct{}

func (cliElectorate) IsIDTokenHolder(a core.Address) bool {
    // Defer to AuthoritySet if available, else allow (dev‑mode).
    if set := core.CurrentAuthoritySet(); set != nil {
        return set.IsAuthority(a)
    }
    return true // permissive fallback for local testing
}

func ensureCharityInitialised(cmd *cobra.Command, _ []string) error {
    if cp != nil {
        return nil // already ready
    }
    led := core.CurrentLedger()
    if led == nil {
        return errors.New("ledger not initialised – start node or init ledger first")
    }
    // Parse genesis timestamp – default to 2025‑01‑01 UTC if env not provided.
    gstr := viper.GetString("CHARITY_GENESIS")
    if gstr == "" {
        gstr = "2025-01-01T00:00:00Z"
    }
    t, err := time.Parse(time.RFC3339, gstr)
    if err != nil {
        return fmt.Errorf("invalid CHARITY_GENESIS env: %w", err)
    }
    genesisTs = t
    cp = core.NewCharityPool(logrus.StandardLogger(), led, cliElectorate{}, genesisTs)
    return nil
}

//---------------------------------------------------------------------
// Controller – façade used by CLI commands
//---------------------------------------------------------------------

type CharityController struct{}

func (c *CharityController) Register(addr core.Address, catStr, name string) error {
    cat, err := parseCategory(catStr)
    if err != nil { return err }
    return cp.Register(addr, name, cat)
}

func (c *CharityController) Vote(voter, charity core.Address) error { return cp.Vote(voter, charity) }

func (c *CharityController) Tick(now time.Time) {
    cp.Tick(now)
}

func (c *CharityController) Winners(cycle uint64) ([]core.Address, error) {
    key := []byte(fmt.Sprintf("charity:winners:%d", cycle))
    raw, _ := core.CurrentLedger().GetState(key)
    if len(raw) == 0 {
        return nil, fmt.Errorf("no winner list for cycle %d", cycle)
    }
    var list []core.Address
    _ = json.Unmarshal(raw, &list)
    return list, nil
}

//---------------------------------------------------------------------
// Helpers
//---------------------------------------------------------------------

func parseCategory(s string) (core.CharityCategory, error) {
    s = strings.TrimSpace(strings.ToLower(s))
    switch s {
    case "hunger", "hungerrelief":
        return core.HungerRelief, nil
    case "children", "childrenhelp":
        return core.ChildrenHelp, nil
    case "wildlife", "wildlifehelp":
        return core.WildlifeHelp, nil
    case "sea", "seasupport":
        return core.SeaSupport, nil
    case "disaster", "disastersupport":
        return core.DisasterSupport, nil
    case "war", "warsupport":
        return core.WarSupport, nil
    default:
        return 0, fmt.Errorf("unknown category %q", s)
    }
}

func currentCycle(now time.Time) uint64 {
    if now.Before(genesisTs) { return 0 }
    return uint64(now.Sub(genesisTs) / cycleDuration)
}

//---------------------------------------------------------------------
// CLI command declarations
//---------------------------------------------------------------------

var charityCmd = &cobra.Command{
    Use:               "charity",
    Short:             "On‑chain charity pool operations (register, vote, payouts)",
    PersistentPreRunE: ensureCharityInitialised,
}

// register -------------------------------------------------------------------
var registerCmd = &cobra.Command{
    Use:   "register <addr> <category> <name>",
    Short: "Register a charity (must occur ≥30 days before cycle end)",
    Args:  cobra.MinimumNArgs(3),
    RunE: func(cmd *cobra.Command, args []string) error {
        ctrl := &CharityController{}
        addr := core.Address(args[0])
        catStr := args[1]
        name := strings.Join(args[2:], " ")
        if err := ctrl.Register(addr, catStr, name); err != nil {
            return err
        }
        fmt.Printf("Charity %s registered in category %s\n", addr, catStr)
        return nil
    },
}

// vote -----------------------------------------------------------------------
var voteCmd = &cobra.Command{
    Use:   "vote <voterAddr> <charityAddr>",
    Short: "Cast a vote (ID‑token holders only) during last 15d of cycle",
    Args:  cobra.ExactArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
        ctrl := &CharityController{}
        voter, charity := core.Address(args[0]), core.Address(args[1])
        if err := ctrl.Vote(voter, charity); err != nil {
            return err
        }
        fmt.Printf("Vote recorded %s → %s\n", voter, charity)
        return nil
    },
}

// tick -----------------------------------------------------------------------
var tickCmd = &cobra.Command{
    Use:   "tick [timestamp RFC3339]",
    Short: "Manually trigger pool cron (daily payout & cycle finalise)",
    Args:  cobra.RangeArgs(0, 1),
    RunE: func(cmd *cobra.Command, args []string) error {
        ctrl := &CharityController{}
        now := time.Now().UTC()
        if len(args) == 1 {
            t, err := time.Parse(time.RFC3339, args[0])
            if err != nil { return fmt.Errorf("invalid timestamp: %w", err) }
            now = t.UTC()
        }
        ctrl.Tick(now)
        fmt.Printf("Tick executed at %s\n", now.Format(time.RFC3339))
        return nil
    },
}

// winners --------------------------------------------------------------------
var winnersCmd = &cobra.Command{
    Use:   "winners [cycle]",
    Short: "Show winner charities list for a cycle (default current)",
    Args:  cobra.RangeArgs(0, 1),
    RunE: func(cmd *cobra.Command, args []string) error {
        ctrl := &CharityController{}
        cycle := currentCycle(time.Now().UTC())
        if len(args) == 1 {
            c, err := strconv.ParseUint(args[0], 10, 64)
            if err != nil { return fmt.Errorf("cycle uint64: %w", err) }
            cycle = c
        }
        winners, err := ctrl.Winners(cycle)
        if err != nil { return err }
        b, _ := json.MarshalIndent(winners, "", "  ")
        fmt.Printf("Winners cycle %d:\n%s\n", cycle, string(b))
        return nil
    },
}

//---------------------------------------------------------------------
// Consolidation & export
//---------------------------------------------------------------------

func init() {
    charityCmd.AddCommand(registerCmd)
    charityCmd.AddCommand(voteCmd)
    charityCmd.AddCommand(tickCmd)
    charityCmd.AddCommand(winnersCmd)
}

// Export for root‑CLI import
var CharityCmd = charityCmd
