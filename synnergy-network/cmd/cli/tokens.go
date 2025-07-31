package cli

// ──────────────────────────────────────────────────────────────────────────────
// Synthron Tokens CLI – inspect & administer on‑chain assets (collision‑free)
// ──────────────────────────────────────────────────────────────────────────────
// Root group   : `tokens`
// Micro‑routes : list, info, balance, transfer, mint, burn, approve, allowance
// All handler / command identifiers are uniquely prefixed with **tok*** to avoid
// name clashes with other CLI modules that expose generic symbols like listCmd.
// ──────────────────────────────────────────────────────────────────────────────

import (
    "encoding/hex"
    "fmt"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/joho/godotenv"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"

    "synnergy-network/core"
)

// -----------------------------------------------------------------------------
// Globals & middleware (runs once)
// -----------------------------------------------------------------------------

var (
    tokLedger *core.Ledger
    tokLogger = logrus.StandardLogger()
    tokOnce   sync.Once
)

func tokInitMiddleware(cmd *cobra.Command, _ []string) error {
    var err error
    tokOnce.Do(func() {
        _ = godotenv.Load()
        lvl := os.Getenv("LOG_LEVEL"); if lvl == "" { lvl = "info" }
        lv, e := logrus.ParseLevel(lvl); if e != nil { err = e; return }
        tokLogger.SetLevel(lv)

        path := os.Getenv("LEDGER_PATH"); if path == "" { err = fmt.Errorf("LEDGER_PATH not set"); return }
        tokLedger, e = core.OpenLedger(path); if e != nil { err = e }
    })
    return err
}

// -----------------------------------------------------------------------------
// Helper utilities
// -----------------------------------------------------------------------------

func tokResolveToken(idOrSym string) (*core.BaseToken, error) {
    // by symbol
    for _, t := range core.GetRegistryTokens() {
        if strings.EqualFold(t.Meta().Symbol, idOrSym) { return t, nil }
    }
    // by id (hex or decimal)
    if strings.HasPrefix(idOrSym, "0x") {
        n, err := strconv.ParseUint(idOrSym[2:], 16, 32); if err != nil { return nil, err }
        tok, ok := core.GetToken(core.TokenID(n)); if !ok { return nil, core.ErrInvalidAsset }; return tok.(*core.BaseToken), nil
    }
    n, err := strconv.ParseUint(idOrSym, 10, 32); if err != nil { return nil, err }
    tok, ok := core.GetToken(core.TokenID(n)); if !ok { return nil, core.ErrInvalidAsset }; return tok.(*core.BaseToken), nil
}

func tokParseAddr(h string) (core.Address, error) {
    var a core.Address
    b, err := hex.DecodeString(strings.TrimPrefix(h, "0x")); if err != nil || len(b) != len(a) { return a, fmt.Errorf("bad address") }
    copy(a[:], b); return a, nil
}

// -----------------------------------------------------------------------------
// Controllers (prefixed names)
// -----------------------------------------------------------------------------

func tokHandleList(cmd *cobra.Command, _ []string) error {
    for _, t := range core.GetRegistryTokens() {
        m := t.Meta(); fmt.Fprintf(cmd.OutOrStdout(), "%d\t%s\t%s\t%d\t%d\n", t.ID(), m.Symbol, m.Name, m.Decimals, m.TotalSupply)
    }
    return nil
}

func tokHandleInfo(cmd *cobra.Command, args []string) error {
    tok, err := tokResolveToken(args[0]); if err != nil { return err }
    m := tok.Meta(); fmt.Fprintf(cmd.OutOrStdout(), "ID: %d\nSymbol: %s\nName: %s\nDecimals: %d\nStandard: 0x%X\nTotalSupply: %d\nCreated: %s\nFixed: %v\n", tok.ID(), m.Symbol, m.Name, m.Decimals, m.Standard, m.TotalSupply, m.Created.Format(time.RFC3339), m.FixedSupply)
    return nil
}

func tokHandleBalance(cmd *cobra.Command, args []string) error {
    tok, err := tokResolveToken(args[0]); if err != nil { return err }
    addr, err := tokParseAddr(args[1]); if err != nil { return err }
    fmt.Fprintf(cmd.OutOrStdout(), "%d\n", tok.BalanceOf(addr)); return nil
}

func tokHandleTransfer(cmd *cobra.Command, args []string) error {
    tok, err := tokResolveToken(args[0]); if err != nil { return err }
    fromStr, _ := cmd.Flags().GetString("from"); toStr, _ := cmd.Flags().GetString("to"); amt, _ := cmd.Flags().GetUint64("amt")
    if fromStr==""||toStr=="" { return fmt.Errorf("--from and --to required") }
    from, err := tokParseAddr(fromStr); if err != nil { return err }
    to, err := tokParseAddr(toStr); if err != nil { return err }
    if err := tok.Transfer(from, to, amt); err != nil { return err }
    fmt.Fprintln(cmd.OutOrStdout(), "transfer ok ✔"); return nil
}

func tokHandleMint(cmd *cobra.Command, args []string) error {
    tok, err := tokResolveToken(args[0]); if err != nil { return err }
    toStr, _ := cmd.Flags().GetString("to"); amt, _ := cmd.Flags().GetUint64("amt")
    to, err := tokParseAddr(toStr); if err != nil { return err }
    if err := tok.Mint(to, amt); err != nil { return err }
    fmt.Fprintln(cmd.OutOrStdout(), "mint ok ✔"); return nil
}

func tokHandleBurn(cmd *cobra.Command, args []string) error {
    tok, err := tokResolveToken(args[0]); if err != nil { return err }
    fromStr, _ := cmd.Flags().GetString("from"); amt, _ := cmd.Flags().GetUint64("amt")
    from, err := tokParseAddr(fromStr); if err != nil { return err }
    if err := tok.Burn(from, amt); err != nil { return err }
    fmt.Fprintln(cmd.OutOrStdout(), "burn ok ✔"); return nil
}

func tokHandleApprove(cmd *cobra.Command, args []string) error {
    tok, err := tokResolveToken(args[0]); if err != nil { return err }
    ownerStr,_:=cmd.Flags().GetString("owner"); spenderStr,_:=cmd.Flags().GetString("spender"); amt,_:=cmd.Flags().GetUint64("amt")
    owner, err := tokParseAddr(ownerStr); if err != nil { return err }
    spender, err := tokParseAddr(spenderStr); if err != nil { return err }
    if err := tok.Approve(owner, spender, amt); err != nil { return err }
    fmt.Fprintln(cmd.OutOrStdout(), "approve ok ✔"); return nil
}

func tokHandleAllowance(cmd *cobra.Command, args []string) error {
    tok, err := tokResolveToken(args[0]); if err != nil { return err }
    owner, err := tokParseAddr(args[1]); if err != nil { return err }
    spender, err := tokParseAddr(args[2]); if err != nil { return err }
    fmt.Fprintf(cmd.OutOrStdout(), "%d\n", tok.Allowance(owner, spender)); return nil
}

// -----------------------------------------------------------------------------
// Cobra command tree (tok‑prefixed vars)
// -----------------------------------------------------------------------------

var tokensCmd = &cobra.Command{
    Use:               "tokens",
    Short:             "Inspect and administer Synthron tokens",
    PersistentPreRunE: tokInitMiddleware,
}

var tokListCmd      = &cobra.Command{Use: "list", Short: "List tokens", Args: cobra.NoArgs, RunE: tokHandleList}
var tokInfoCmd      = &cobra.Command{Use: "info <id|symbol>", Short: "Token metadata", Args: cobra.ExactArgs(1), RunE: tokHandleInfo}
var tokBalCmd       = &cobra.Command{Use: "balance <tok> <addr>", Short: "Balance", Args: cobra.ExactArgs(2), RunE: tokHandleBalance}
var tokTransferCmd  = &cobra.Command{Use: "transfer <tok>", Short: "Transfer", Args: cobra.ExactArgs(1), RunE: tokHandleTransfer}
var tokMintCmd      = &cobra.Command{Use: "mint <tok>", Short: "Mint (admin)", Args: cobra.ExactArgs(1), RunE: tokHandleMint}
var tokBurnCmd      = &cobra.Command{Use: "burn <tok>", Short: "Burn (admin)", Args: cobra.ExactArgs(1), RunE: tokHandleBurn}
var tokApproveCmd   = &cobra.Command{Use: "approve <tok>", Short: "Approve spender", Args: cobra.ExactArgs(1), RunE: tokHandleApprove}
var tokAllowanceCmd = &cobra.Command{Use: "allowance <tok> <owner> <spender>", Short: "Allowance", Args: cobra.ExactArgs(3), RunE: tokHandleAllowance}

func init() {
    // transfer flags
    tokTransferCmd.Flags().String("from", "", "sender address")
    tokTransferCmd.Flags().String("to", "", "recipient address")
    tokTransferCmd.Flags().Uint64("amt", 0, "amount")
    tokTransferCmd.MarkFlagRequired("from"); tokTransferCmd.MarkFlagRequired("to"); tokTransferCmd.MarkFlagRequired("amt")

    // mint flags
    tokMintCmd.Flags().String("to", "", "recipient address"); tokMintCmd.Flags().Uint64("amt", 0, "amount")
    tokMintCmd.MarkFlagRequired("to"); tokMintCmd.MarkFlagRequired("amt")

    // burn flags
    tokBurnCmd.Flags().String("from", "", "holder"); tokBurnCmd.Flags().Uint64("amt", 0, "amount")
    tokBurnCmd.MarkFlagRequired("from"); tokBurnCmd.MarkFlagRequired("amt")

    // approve flags
    tokApproveCmd.Flags().String("owner", "", "owner address")
    tokApproveCmd.Flags().String("spender", "", "spender address")
    tokApproveCmd.Flags().Uint64("amt", 0, "amount")
    tokApproveCmd.MarkFlagRequired("owner"); tokApproveCmd.MarkFlagRequired("spender"); tokApproveCmd.MarkFlagRequired("amt")

    // attach sub‑commands to root
    tokensCmd.AddCommand(tokListCmd, tokInfoCmd, tokBalCmd, tokTransferCmd, tokMintCmd, tokBurnCmd, tokApproveCmd, tokAllowanceCmd)
}

// -----------------------------------------------------------------------------
// Consolidated export
// -----------------------------------------------------------------------------

var TokensCmd = tokensCmd

func RegisterTokens(root *cobra.Command) { root.AddCommand(TokensCmd) }
