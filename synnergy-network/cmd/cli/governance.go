// cmd/cli/governance.go – Governance/DAO CLI for Synnergy Network
// -----------------------------------------------------------------------------
// This file exposes a consolidated “~gov” (aka "~governance") command grouping
// all governance‑related sub‑commands. It follows the same design contract as
// the fault‑tolerance CLI:
// • All top‑level Cobra commands are declared first for readability.
// • Controllers wrap JSON‑RPC calls to the governance daemon (newline‑framed).
// • Routes get wired up in a single init() block at the bottom, and the public
//   factory NewGovernanceCommand() returns the root for easy import into the
//   global CLI index.
// -----------------------------------------------------------------------------
// Example usage
//   synnergy ~gov propose --changes="block_gas_limit=1500000" --desc="Raise gas" --deadline=48h
//   synnergy ~gov vote <proposal‑id> --approve
//   synnergy ~gov execute <proposal‑id>
//   synnergy ~gov list --format=json
// -----------------------------------------------------------------------------
// Environment / Config
//   GOVERNANCE_API_ADDR – host:port of governance daemon (default "127.0.0.1:7700")
// -----------------------------------------------------------------------------

package cli

import (
    "bufio"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "net"
    "os"
    "strings"
    "time"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"

    core "synnergy-network/core"
)

// -----------------------------------------------------------------------------
// Middleware – JSON/TCP RPC client
// -----------------------------------------------------------------------------

type govClient struct {
    conn net.Conn
    rd   *bufio.Reader
}

func newGovClient(ctx context.Context) (*govClient, error) {
    addr := viper.GetString("GOVERNANCE_API_ADDR")
    if addr == "" {
        addr = "127.0.0.1:7700"
    }
    d := net.Dialer{}
    conn, err := d.DialContext(ctx, "tcp", addr)
    if err != nil {
        return nil, fmt.Errorf("cannot connect to governance daemon at %s: %w", addr, err)
    }
    return &govClient{conn: conn, rd: bufio.NewReader(conn)}, nil
}

func (c *govClient) Close() { _ = c.conn.Close() }

func (c *govClient) writeJSON(v any) error {
    b, err := json.Marshal(v)
    if err != nil {
        return err
    }
    b = append(b, '\n')
    _, err = c.conn.Write(b)
    return err
}

func (c *govClient) readJSON(v any) error {
    dec := json.NewDecoder(c.rd)
    return dec.Decode(v)
}

// -----------------------------------------------------------------------------
// Controller helpers – wrap RPC calls
// -----------------------------------------------------------------------------

func submitProposal(ctx context.Context, changes map[string]string, desc string, deadline time.Duration) (string, error) {
    cli, err := newGovClient(ctx)
    if err != nil {
        return "", err
    }
    defer cli.Close()

    payload := map[string]any{
        "action":      "propose",
        "changes":     changes,
        "description": desc,
        "deadline":    int64(deadline.Seconds()), // secs from now
    }
    if err := cli.writeJSON(payload); err != nil {
        return "", err
    }
    var resp struct {
        ID    string `json:"id"`
        Error string `json:"error,omitempty"`
    }
    if err := cli.readJSON(&resp); err != nil {
        return "", err
    }
    if resp.Error != "" {
        return "", errors.New(resp.Error)
    }
    return resp.ID, nil
}

func castVote(ctx context.Context, id string, approve bool) error {
    cli, err := newGovClient(ctx)
    if err != nil {
        return err
    }
    defer cli.Close()
    return cli.writeJSON(map[string]any{"action": "vote", "id": id, "approve": approve})
}

func executeProposal(ctx context.Context, id string) error {
    cli, err := newGovClient(ctx)
    if err != nil {
        return err
    }
    defer cli.Close()
    return cli.writeJSON(map[string]any{"action": "execute", "id": id})
}

func listProposals(ctx context.Context) ([]core.GovProposal, error) {
    cli, err := newGovClient(ctx)
    if err != nil {
        return nil, err
    }
    defer cli.Close()
    if err := cli.writeJSON(map[string]any{"action": "list"}); err != nil {
        return nil, err
    }
    var resp struct {
        Proposals []core.GovProposal `json:"proposals"`
        Error     string            `json:"error,omitempty"`
    }
    if err := cli.readJSON(&resp); err != nil {
        return nil, err
    }
    if resp.Error != "" {
        return nil, errors.New(resp.Error)
    }
    return resp.Proposals, nil
}

func getProposal(ctx context.Context, id string) (*core.GovProposal, error) {
    cli, err := newGovClient(ctx)
    if err != nil {
        return nil, err
    }
    defer cli.Close()
    if err := cli.writeJSON(map[string]any{"action": "get", "id": id}); err != nil {
        return nil, err
    }
    var resp struct {
        Proposal core.GovProposal `json:"proposal"`
        Error    string          `json:"error,omitempty"`
    }
    if err := cli.readJSON(&resp); err != nil {
        return nil, err
    }
    if resp.Error != "" {
        return nil, errors.New(resp.Error)
    }
    return &resp.Proposal, nil
}

// -----------------------------------------------------------------------------
// Top‑level CLI commands
// -----------------------------------------------------------------------------

var govCmd = &cobra.Command{
    Use:     "~gov",
    Short:   "Network governance & DAO operations",
    Aliases: []string{"gov", "governance"},
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        cobra.OnInitialize(initGovConfig)
        return nil
    },
}

// propose ---------------------------------------------------------------------
var proposeCmd = &cobra.Command{
    Use:   "propose",
    Short: "Submit a new proposal",
    RunE: func(cmd *cobra.Command, args []string) error {
        changesStr, _ := cmd.Flags().GetString("changes")
        desc, _ := cmd.Flags().GetString("desc")
        dlStr, _ := cmd.Flags().GetString("deadline")

        if changesStr == "" {
            return errors.New("--changes required")
        }
        changes := map[string]string{}
        for _, pair := range strings.Split(changesStr, ",") {
            kv := strings.SplitN(pair, "=", 2)
            if len(kv) != 2 {
                return fmt.Errorf("invalid change pair %q", pair)
            }
            changes[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
        }

        deadline := 72 * time.Hour // default 3d
        if dlStr != "" {
            d, err := time.ParseDuration(dlStr)
            if err != nil {
                return fmt.Errorf("invalid --deadline: %w", err)
            }
            deadline = d
        }

        ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
        defer cancel()
        id, err := submitProposal(ctx, changes, desc, deadline)
        if err != nil {
            return err
        }
        fmt.Printf("Proposal submitted: %s\n", id)
        return nil
    },
}

// vote ------------------------------------------------------------------------
var voteCmd = &cobra.Command{
    Use:   "vote [proposal-id]",
    Short: "Cast a vote on a proposal",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        approve, _ := cmd.Flags().GetBool("approve")
        ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
        defer cancel()
        return castVote(ctx, args[0], approve)
    },
}

// execute ---------------------------------------------------------------------
var execCmd = &cobra.Command{
    Use:   "execute [proposal-id]",
    Short: "Execute a proposal after the deadline",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
        defer cancel()
        return executeProposal(ctx, args[0])
    },
}

// get -------------------------------------------------------------------------
var getCmd = &cobra.Command{
    Use:   "get [proposal-id]",
    Short: "Show a single proposal",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
        defer cancel()
        p, err := getProposal(ctx, args[0])
        if err != nil {
            return err
        }
        enc := json.NewEncoder(os.Stdout)
        enc.SetIndent("", "  ")
        return enc.Encode(p)
    },
}

// list ------------------------------------------------------------------------
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List all proposals",
    RunE: func(cmd *cobra.Command, args []string) error {
        ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
        defer cancel()
        props, err := listProposals(ctx)
        if err != nil {
            return err
        }
        switch viper.GetString("output.format") {
        case "json":
            enc := json.NewEncoder(os.Stdout)
            enc.SetIndent("", "  ")
            return enc.Encode(props)
        default:
            fmt.Printf("%-36s %-20s %-8s %-8s %-5s %s\n", "ID", "Created", "For", "Against", "Exec", "Description")
            for _, p := range props {
                execFlag := "no"
                if p.Executed {
                    execFlag = "yes"
                }
                fmt.Printf("%-36s %-20s %-8d %-8d %-5s %s\n", p.ID, p.Created.Format(time.RFC3339), p.VotesFor, p.VotesAgainst, execFlag, p.Description)
            }
            return nil
        }
    },
}

// -----------------------------------------------------------------------------
// init – wiring routes & flags
// -----------------------------------------------------------------------------

func initGovConfig() {
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

    viper.SetDefault("GOVERNANCE_API_ADDR", "127.0.0.1:7700")
    viper.SetDefault("output.format", "table")
}

func init() {
    proposeCmd.Flags().String("changes", "", "comma‑separated key=value list of param changes")
    proposeCmd.Flags().String("desc", "", "human‑readable description")
    proposeCmd.Flags().String("deadline", "", "voting deadline as duration (e.g. 48h), default 72h")

    voteCmd.Flags().Bool("approve", true, "vote approve=true / deny=false")

    listCmd.Flags().StringP("format", "f", "table", "output format: table|json")
    _ = viper.BindPFlag("output.format", listCmd.Flags().Lookup("format"))

    // Register sub‑commands
    govCmd.AddCommand(proposeCmd)
    govCmd.AddCommand(voteCmd)
    govCmd.AddCommand(execCmd)
    govCmd.AddCommand(getCmd)
    govCmd.AddCommand(listCmd)
}

// NewGovernanceCommand returns a consolidated Cobra command tree for the root
// CLI index (e.g. rootCmd.AddCommand(cli.NewGovernanceCommand())).
func NewGovernanceCommand() *cobra.Command { return govCmd }
