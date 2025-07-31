// cmd/cli/replication.go – Block‑replication & sync CLI
// -----------------------------------------------------------------------------
// Provides operational control over the replication subsystem via the unified
// route “~rep”.  All commands rely on a newline‑framed JSON‑RPC control socket
// exposed by the replication daemon.
//
// Top‑level commands (declared first):
//   • start        – launch replication loops (idempotent)
//   • stop         – terminate replication loops gracefully
//   • status       – show peer/queue stats
//   • replicate    – manually gossip a known block hash
//   • request      – fetch an absent block by hash
//
// Route wiring occurs in the single init() block at the bottom; public factory
// NewReplicationCommand() returns the consolidated Cobra tree.
// -----------------------------------------------------------------------------
// Examples
//   synnergy ~rep start
//   synnergy ~rep status --format=json
//   synnergy ~rep replicate deadbeef…cafebabe
//   synnergy ~rep request 0123…89ab
// -----------------------------------------------------------------------------
// Environment
//   REPL_API_ADDR – host:port of replication daemon (default "127.0.0.1:7950")
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
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// -----------------------------------------------------------------------------
// Middleware – thin framed JSON/TCP client
// -----------------------------------------------------------------------------

type replClient struct {
	conn net.Conn
	rd   *bufio.Reader
}

func newReplClient(ctx context.Context) (*replClient, error) {
	addr := viper.GetString("REPL_API_ADDR")
	if addr == "" {
		addr = "127.0.0.1:7950"
	}
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to replication daemon at %s: %w", addr, err)
	}
	return &replClient{conn: conn, rd: bufio.NewReader(conn)}, nil
}

func (c *replClient) Close() { _ = c.conn.Close() }

func (c *replClient) writeJSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = c.conn.Write(b)
	return err
}

func (c *replClient) readJSON(v any) error {
	dec := json.NewDecoder(c.rd)
	return dec.Decode(v)
}

// -----------------------------------------------------------------------------
// Controller helpers – RPC entry‑points
// -----------------------------------------------------------------------------

func startRPC(ctx context.Context) error {
	cli, err := newReplClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "start"})
}

func stopRPC(ctx context.Context) error {
	cli, err := newReplClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "stop"})
}

func statusRPC(ctx context.Context) (map[string]any, error) {
	cli, err := newReplClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "status"}); err != nil {
		return nil, err
	}
	var resp struct {
		Data  map[string]any `json:"data"`
		Error string         `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return resp.Data, nil
}

func replicateRPC(ctx context.Context, hashHex string) error {
	cli, err := newReplClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "replicate", "hash": hashHex})
}

func requestRPC(ctx context.Context, hashHex string) error {
	cli, err := newReplClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	// this endpoint returns error only; successful fetch prints via daemon logs
	return cli.writeJSON(map[string]any{"action": "request", "hash": hashHex})
}

func syncRPC(ctx context.Context) error {
	cli, err := newReplClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "sync"})
}

// -----------------------------------------------------------------------------
// Top‑level Cobra commands
// -----------------------------------------------------------------------------

var repCmd = &cobra.Command{
	Use:     "~rep",
	Short:   "Block replication & fast‑sync control",
	Aliases: []string{"rep", "replication"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cobra.OnInitialize(initReplConfig)
		return nil
	},
}

// start -----------------------------------------------------------------------
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Launch replication goroutines (idempotent)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		return startRPC(ctx)
	},
}

// stop ------------------------------------------------------------------------
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop replication goroutines gracefully",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		return stopRPC(ctx)
	},
}

// status ----------------------------------------------------------------------
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show replication subsystem status",
	RunE: func(cmd *cobra.Command, args []string) error {
		format := viper.GetString("output.format")
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		data, err := statusRPC(ctx)
		if err != nil {
			return err
		}
		switch format {
		case "json":
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(data)
		default:
			for k, v := range data {
				fmt.Printf("%s: %v\n", k, v)
			}
			return nil
		}
	},
}

// replicate -------------------------------------------------------------------
var replicateCmd = &cobra.Command{
	Use:   "replicate [block‑hash]",
	Short: "Manually gossip an already‑known block",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := hex.DecodeString(args[0]); err != nil || len(args[0]) != 64 {
			return errors.New("block‑hash must be 32‑byte hex string")
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		return replicateRPC(ctx, args[0])
	},
}

// request ---------------------------------------------------------------------
var requestCmd = &cobra.Command{
	Use:   "request [block‑hash]",
	Short: "Fetch a missing block from peers",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := hex.DecodeString(args[0]); err != nil || len(args[0]) != 64 {
			return errors.New("block‑hash must be 32‑byte hex string")
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		return requestRPC(ctx, args[0])
	},
}

// sync -----------------------------------------------------------------------
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize blocks from peers",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		return syncRPC(ctx)
	},
}

// -----------------------------------------------------------------------------
// init – config bootstrap & route registration
// -----------------------------------------------------------------------------

func initReplConfig() {
	viper.SetEnvPrefix("synnergy")
	viper.AutomaticEnv()

	cfgFile := viper.GetString("config")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("synnergy")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config/synnergy")
	}
	_ = viper.ReadInConfig()

	viper.SetDefault("REPL_API_ADDR", "127.0.0.1:7950")
	viper.SetDefault("output.format", "table")
}

func init() {
	// flag binding for status output format
	statusCmd.Flags().StringP("format", "f", "table", "output format: table|json")
	_ = viper.BindPFlag("output.format", statusCmd.Flags().Lookup("format"))

	// sub‑command registration
	repCmd.AddCommand(startCmd)
	repCmd.AddCommand(stopCmd)
	repCmd.AddCommand(statusCmd)
	repCmd.AddCommand(replicateCmd)
	repCmd.AddCommand(requestCmd)
	repCmd.AddCommand(syncCmd)
}

// NewReplicationCommand returns the root Cobra command for ~rep.
func NewReplicationCommand() *cobra.Command { return repCmd }
