// cmd/cli/sidechain.go – Side‑chain bridge & header‑sync CLI
// -----------------------------------------------------------------------------
// Consolidated under “~sc” (aka “~sidechain”).  Wraps all high‑level bridge
// operations and pipes them to the side‑chain coordinator daemon via newline
// framed JSON‑RPC.
// -----------------------------------------------------------------------------
// Commands
//   register      – register a new side‑chain (gov/admin only)
//   header        – submit finalised side‑chain header
//   deposit       – deposit tokens into a side‑chain escrow
//   withdraw      – verify L2→L1 withdrawal proof
//   meta          – show side‑chain metadata
//   list          – list all registered side‑chains
// -----------------------------------------------------------------------------
// Environment / Config
//   SIDECHAIN_API_ADDR – host:port of coordinator daemon (default "127.0.0.1:7990")
// -----------------------------------------------------------------------------

package cli

import (
    "bufio"
    "context"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "net"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

// -----------------------------------------------------------------------------
// Middleware – framed JSON/TCP client
// -----------------------------------------------------------------------------

type scClient struct {
    conn net.Conn
    rd   *bufio.Reader
}

func newSCClient(ctx context.Context) (*scClient, error) {
    addr := viper.GetString("SIDECHAIN_API_ADDR")
    if addr == "" { addr = "127.0.0.1:7990" }
    d := net.Dialer{}
    conn, err := d.DialContext(ctx, "tcp", addr)
    if err != nil { return nil, fmt.Errorf("cannot connect to sidechain daemon at %s: %w", addr, err) }
    return &scClient{conn: conn, rd: bufio.NewReader(conn)}, nil
}

func (c *scClient) Close() { _ = c.conn.Close() }

func (c *scClient) writeJSON(v any) error {
    b, err := json.Marshal(v)
    if err != nil { return err }
    b = append(b, '\n')
    _, err = c.conn.Write(b)
    return err
}

func (c *scClient) readJSON(v any) error {
    dec := json.NewDecoder(c.rd)
    return dec.Decode(v)
}

// -----------------------------------------------------------------------------
// Controller helpers
// -----------------------------------------------------------------------------

func registerRPC(ctx context.Context, id uint32, name string, threshold uint8, vals []string) error {
    cli, err := newSCClient(ctx)
    if err != nil { return err }
    defer cli.Close()
    return cli.writeJSON(map[string]any{"action": "register", "id": id, "name": name, "threshold": threshold, "validators": vals})
}

func headerRPC(ctx context.Context, payload map[string]any) error {
    cli, err := newSCClient(ctx)
    if err != nil { return err }
    defer cli.Close()
    return cli.writeJSON(payload)
}

func depositRPC(ctx context.Context, chain uint32, from string, to string, token string, amt uint64) (uint64, error) {
    cli, err := newSCClient(ctx)
    if err != nil { return 0, err }
    defer cli.Close()
    if err := cli.writeJSON(map[string]any{"action": "deposit", "chain": chain, "from": from, "to": to, "token": token, "amount": amt}); err != nil { return 0, err }
    var resp struct{ Nonce uint64 `json:"nonce"`; Error string `json:"error,omitempty"` }
    if err := cli.readJSON(&resp); err != nil { return 0, err }
    if resp.Error != "" { return 0, errors.New(resp.Error) }
    return resp.Nonce, nil
}

func withdrawRPC(ctx context.Context, proofHex string) error {
    cli, err := newSCClient(ctx)
    if err != nil { return err }
    defer cli.Close()
    return cli.writeJSON(map[string]any{"action": "withdraw", "proof": proofHex})
}

func metaRPC(ctx context.Context, id uint32) (map[string]any, error) {
    cli, err := newSCClient(ctx)
    if err != nil { return nil, err }
    defer cli.Close()
    if err := cli.writeJSON(map[string]any{"action": "meta", "id": id}); err != nil { return nil, err }
    var resp struct{ Meta map[string]any `json:"meta"`; Error string `json:"error,omitempty"` }
    if err := cli.readJSON(&resp); err != nil { return nil, err }
    if resp.Error != "" { return nil, errors.New(resp.Error) }
    return resp.Meta, nil
}

func listRPC(ctx context.Context) ([]map[string]any, error) {
    cli, err := newSCClient(ctx)
    if err != nil { return nil, err }
    defer cli.Close()
    if err := cli.writeJSON(map[string]any{"action": "list"}); err != nil { return nil, err }
    var resp struct{ List []map[string]any `json:"list"`; Error string `json:"error,omitempty"` }
    if err := cli.readJSON(&resp); err != nil { return nil, err }
    if resp.Error != "" { return nil, errors.New(resp.Error) }
    return resp.List, nil
}

// -----------------------------------------------------------------------------
// Top-level cobra tree
// -----------------------------------------------------------------------------

var scCmd = &cobra.Command{
    Use:     "~sc",
    Short:   "Side‑chain bridge operations",
    Aliases: []string{"sc", "sidechain"},
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        cobra.OnInitialize(initSCConfig)
        return nil
    },
}

// register --------------------------------------------------------------------
var registerCmd = &cobra.Command{
    Use:   "register",
    Short: "Register new side‑chain (governance only)",
    RunE: func(cmd *cobra.Command, args []string) error {
        idU, _ := cmd.Flags().GetUint32("id")
        name, _ := cmd.Flags().GetString("name")
        thrU, _ := cmd.Flags().GetUint("threshold")
        valsStr, _ := cmd.Flags().GetString("validators")
        if name == "" || valsStr == "" { return errors.New("--name and --validators required") }
        vals := strings.Split(valsStr, ",")
        for _, v := range vals {
            if _, err := hex.DecodeString(v); err != nil { return fmt.Errorf("invalid validator pubkey: %w", err) }
        }
        ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
        defer cancel()
        return registerRPC(ctx, idU, name, uint8(thrU), vals)
    },
}

// header ----------------------------------------------------------------------
var headerCmd = &cobra.Command{
    Use:   "header",
    Short: "Submit side‑chain header (JSON via --file or flags)",
    RunE: func(cmd *cobra.Command, args []string) error {
        file, _ := cmd.Flags().GetString("file")
        var payload map[string]any
        if file != "" {
            f, err := os.ReadFile(file)
            if err != nil { return err }
            if err := json.Unmarshal(f, &payload); err != nil { return err }
        } else {
            // gather required fields via flags
            chain, _ := cmd.Flags().GetUint32("chain")
            height, _ := cmd.Flags().GetUint64("height")
            blockRoot, _ := cmd.Flags().GetString("blockroot")
            txRoot, _ := cmd.Flags().GetString("txroot")
            stateRoot, _ := cmd.Flags().GetString("stateroot")
            sigAgg, _ := cmd.Flags().GetString("sig")
            for _, fld := range []string{blockRoot, txRoot, stateRoot, sigAgg} {
                if fld == "" { return errors.New("missing header field") }
            }
            payload = map[string]any{
                "action":     "header",
                "chain":      chain,
                "height":     height,
                "block_root": blockRoot,
                "tx_root":    txRoot,
                "state_root": stateRoot,
                "sig_agg":    sigAgg,
            }
        }
        if payload["action"] == nil { payload["action"] = "header" }
        ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
        defer cancel()
        return headerRPC(ctx, payload)
    },
}

// deposit ---------------------------------------------------------------------
var depositCmd = &cobra.Command{
    Use:   "deposit",
    Short: "Deposit tokens to side‑chain escrow",
    RunE: func(cmd *cobra.Command, args []string) error {
        chain, _ := cmd.Flags().GetUint32("chain")
        from, _ := cmd.Flags().GetString("from")
        to, _ := cmd.Flags().GetString("to")
        token, _ := cmd.Flags().GetString("token")
        amtU, _ := cmd.Flags().GetUint64("amount")
        for _, fld := range []string{from, to, token} {
            if fld == "" { return errors.New("--from --to --token required") }
        }
        ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
        defer cancel()
        nonce, err := depositRPC(ctx, chain, from, to, token, amtU)
        if err != nil { return err }
        fmt.Printf("deposit nonce: %d\n", nonce)
        return nil
    },
}

// withdraw --------------------------------------------------------------------
var withdrawCmd = &cobra.Command{
    Use:   "withdraw [proofHex]",
    Short: "Verify L2→L1 withdrawal proof",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        proof := args[0]
        if _, err := hex.DecodeString(proof); err != nil { return fmt.Errorf("invalid proof hex: %w", err) }
        ctx, cancel := context.WithTimeout(cmd.Context(), 4*time.Second)
        defer cancel()
        return withdrawRPC(ctx, proof)
    },
}

// meta ------------------------------------------------------------------------
var metaCmd = &cobra.Command{
    Use:   "meta [chainID]",
    Short: "Show side‑chain metadata",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        idU, err := strconv.ParseUint(args[0], 10, 32)
        if err != nil { return fmt.Errorf("invalid chainID: %w", err) }
        ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
        defer cancel()
        meta, err := metaRPC(ctx, uint32(idU))
        if err != nil { return err }
        enc := json.NewEncoder(os.Stdout)
        enc.SetIndent("", "  ")
        return enc.Encode(meta)
    },
}

// list ------------------------------------------------------------------------
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List all registered side‑chains",
    RunE: func(cmd *cobra.Command, args []string) error {
        ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
        defer cancel()
        list, err := listRPC(ctx)
        if err != nil { return err }
        enc := json.NewEncoder(os.Stdout)
        enc.SetIndent("", "  ")
        return enc.Encode(list)
    },
}

// -----------------------------------------------------------------------------
// init – config & route wiring
// -----------------------------------------------------------------------------

func initSCConfig() {
    viper.SetEnvPrefix("synnergy")
    viper.AutomaticEnv()

    cfg := viper.GetString("config")
    if cfg != "" {
        viper.SetConfigFile(cfg)
    } else {
        viper.SetConfigName("synnergy")
        viper.AddConfigPath(".")
        viper.AddConfigPath("$HOME/.config/synnergy")
    }
    _ = viper.ReadInConfig()

    viper.SetDefault("SIDECHAIN_API_ADDR", "127.0.0.1:7990")
}

func init() {
    // register flags
    registerCmd.Flags().Uint32("id", 0, "side‑chain ID")
    registerCmd.Flags().String("name", "", "human name")
    registerCmd.Flags().Uint("threshold", 67, "BLS threshold percent")
    registerCmd.Flags().String("validators", "", "comma‑separated BLS pubkeys hex")

    headerCmd.Flags().String("file", "", "path to header JSON file")
    headerCmd.Flags().Uint32("chain", 0, "chainID")
    headerCmd.Flags().Uint64("height", 0, "header height")
    headerCmd.Flags().String("blockroot", "", "blockRoot hex")
    headerCmd.Flags().String("txroot", "", "txRoot hex")
    headerCmd.Flags().String("stateroot", "", "stateRoot hex")
    headerCmd.Flags().String("sig", "", "aggregate signature hex")

    depositCmd.Flags().Uint32("chain", 0, "chainID")
    depositCmd.Flags().String("from", "", "sender address hex")
    depositCmd.Flags().String("to", "", "L2 recipient bytes hex")
    depositCmd.Flags().String("token", "", "token ID / symbol")
    depositCmd.Flags().Uint64("amount", 0, "amount")

    // route wiring
    scCmd.AddCommand(registerCmd)
    scCmd.AddCommand(headerCmd)
    scCmd.AddCommand(depositCmd)
    scCmd.AddCommand(withdrawCmd)
    scCmd.AddCommand(metaCmd)
    scCmd.AddCommand(listCmd)
}

// NewSidechainCommand exposes the consolidated CLI route.
func NewSidechainCommand() *cobra.Command { return scCmd }
