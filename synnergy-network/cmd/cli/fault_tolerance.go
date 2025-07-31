// cmd/cli/fault_tolerance.go – Production‑grade Cobra CLI for the Synnergy Network
// fault‑tolerance subsystem. Exposes a consolidated “~fault” route under which all
// health‑checking and view‑change operations can be invoked.
// -----------------------------------------------------------------------------
// Usage examples
// --------------
// • List live peer statistics (JSON)
//     synnergy ~fault snapshot --format=json
// • Add a peer to the active health‑checker set
//     synnergy ~fault add-peer 10.0.0.52:9090
// • Remove a peer
//     synnergy ~fault rm-peer 76c2…ffae
// • Trigger a manual view‑change (administrative override)
//     synnergy ~fault view-change --reason="leader unresponsive"
// -----------------------------------------------------------------------------
// Environment
// -----------
// FAULT_API_ADDR – host:port of the fault‑tolerance daemon (default "127.0.0.1:7600")
//
// Add the following to your .env (or any supported config file picked up by
// Viper):
//     FAULT_API_ADDR=127.0.0.1:7600
// -----------------------------------------------------------------------------

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core" // import path adjusted to project layout
)

// -----------------------------------------------------------------------------
// Middleware: config bootstrap & RPC dialer
// -----------------------------------------------------------------------------

type faultClient struct {
	conn net.Conn
}

func newFaultClient(ctx context.Context) (*faultClient, error) {
	addr := viper.GetString("FAULT_API_ADDR")
	if addr == "" {
		addr = "127.0.0.1:7600"
	}
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to fault‑tolerance daemon at %s: %w", addr, err)
	}
	return &faultClient{conn: conn}, nil
}

func (c *faultClient) Close() { _ = c.conn.Close() }

// writeJSON marshals v to JSON with newline delimiter (simple framing) and writes it.
func (c *faultClient) writeJSON(v any) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	bytes = append(bytes, '\n')
	_, err = c.conn.Write(bytes)
	return err
}

// readJSON reads a single JSON value (newline framed) into v.
func (c *faultClient) readJSON(v any) error {
	dec := json.NewDecoder(c.conn)
	return dec.Decode(v)
}

// -----------------------------------------------------------------------------
// Controller helpers (wrap RPC calls)
// -----------------------------------------------------------------------------

func snapshotPeers(ctx context.Context) ([]core.PeerInfo, error) {
	cli, err := newFaultClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "snapshot"}); err != nil {
		return nil, err
	}
	var resp struct {
		Peers []core.PeerInfo `json:"peers"`
		Error string          `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf(resp.Error)
	}
	return resp.Peers, nil
}

func addPeer(ctx context.Context, addr string) error {
	cli, err := newFaultClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "add_peer", "addr": addr})
}

func rmPeer(ctx context.Context, id string) error {
	cli, err := newFaultClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "rm_peer", "addr": id})
}

func manualViewChange(ctx context.Context, reason string) error {
	cli, err := newFaultClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "view_change", "reason": reason})
}

func backupSnapshot(ctx context.Context, incremental bool) error {
	cli, err := newFaultClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "backup", "incremental": incremental})
}

func restoreSnapshot(ctx context.Context, path string) error {
	cli, err := newFaultClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	abs := filepath.Clean(path)
	return cli.writeJSON(map[string]any{"action": "restore", "path": abs})
}

func triggerFailover(ctx context.Context, addr string) error {
	cli, err := newFaultClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "failover", "addr": addr})
}

func predictiveCheck(ctx context.Context, addr string) (float64, error) {
	cli, err := newFaultClient(ctx)
	if err != nil {
		return 0, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "predict", "addr": addr}); err != nil {
		return 0, err
	}
	var resp struct {
		Prob  float64 `json:"prob"`
		Error string  `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return 0, err
	}
	if resp.Error != "" {
		return 0, fmt.Errorf(resp.Error)
	}
	return resp.Prob, nil
}

// -----------------------------------------------------------------------------
// CLI commands (top‑declared for discoverability)
// -----------------------------------------------------------------------------

var faultCmd = &cobra.Command{
	Use:     "~fault",
	Short:   "Fault‑tolerance & health–check operations",
	Aliases: []string{"fault", "ft"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Configuration middleware – load .env & config files only once.
		cobra.OnInitialize(initConfig)
		return nil
	},
}

// snapshot command -------------------------------------------------------------
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Dump current peer statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()

		peers, err := snapshotPeers(ctx)
		if err != nil {
			return err
		}

		switch viper.GetString("output.format") {
		case "json":
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(peers)
		default:
			// table output
			fmt.Printf("%-24s %-6s %-6s %s\n", "Address", "RTT", "Miss", "Updated")
			for _, p := range peers {
				fmt.Printf("%-24s %5.0fms %6d %s\n", p.Address, p.RTT, p.Misses, time.Unix(p.Updated, 0).Format(time.RFC3339))
			}
			return nil
		}
	},
}

// add‑peer command -------------------------------------------------------------
var addPeerCmd = &cobra.Command{
	Use:   "add-peer [addr]",
	Short: "Add peer to health‑checker set",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		return addPeer(ctx, args[0])
	},
}

// rm‑peer command --------------------------------------------------------------
var rmPeerCmd = &cobra.Command{
	Use:   "rm-peer [addr|id]",
	Short: "Remove peer from health‑checker set",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		return rmPeer(ctx, args[0])
	},
}

// view‑change command ----------------------------------------------------------
var viewChangeCmd = &cobra.Command{
	Use:   "view-change",
	Short: "Force a leader rotation",
	RunE: func(cmd *cobra.Command, args []string) error {
		reason, _ := cmd.Flags().GetString("reason")
		if reason == "" {
			reason = "manual override"
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		return manualViewChange(ctx, reason)
	},
}

// backup command --------------------------------------------------------------
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a ledger backup snapshot",
	RunE: func(cmd *cobra.Command, args []string) error {
		inc, _ := cmd.Flags().GetBool("incremental")
		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()
		return backupSnapshot(ctx, inc)
	},
}

// restore command -------------------------------------------------------------
var restoreCmd = &cobra.Command{
	Use:   "restore [file]",
	Short: "Restore ledger state from a snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()
		return restoreSnapshot(ctx, args[0])
	},
}

// failover command ------------------------------------------------------------
var failoverCmd = &cobra.Command{
	Use:   "failover [addr]",
	Short: "Force failover of the given node",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		return triggerFailover(ctx, args[0])
	},
}

// predict command -------------------------------------------------------------
var predictCmd = &cobra.Command{
	Use:   "predict [addr]",
	Short: "Predict failure probability for a node",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		prob, err := predictiveCheck(ctx, args[0])
		if err != nil {
			return err
		}
		fmt.Printf("failure probability: %.2f\n", prob)
		return nil
	},
}

// -----------------------------------------------------------------------------
// initConfig – loads config files, env and default flag wiring.
// -----------------------------------------------------------------------------

func initConfig() {
	viper.SetEnvPrefix("synnergy")
	viper.AutomaticEnv() // env variables take precedence

	// Support common config file locations
	cfgFile := viper.GetString("config")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("synnergy")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config/synnergy")
	}
	_ = viper.ReadInConfig() // nolint – config is optional

	viper.SetDefault("FAULT_API_ADDR", "127.0.0.1:7600")
	viper.SetDefault("output.format", "table")
}

// -----------------------------------------------------------------------------
// Consolidated routing
// -----------------------------------------------------------------------------

func init() {
	// Flag wiring & sub‑command registration happen in init() so callers only
	// need to import the package for side‑effects before attaching the route.

	snapshotCmd.Flags().StringP("format", "f", "table", "output format: table|json")
	_ = viper.BindPFlag("output.format", snapshotCmd.Flags().Lookup("format"))

	viewChangeCmd.Flags().String("reason", "", "reason for manual view‑change")
	backupCmd.Flags().Bool("incremental", false, "incremental snapshot")

	faultCmd.AddCommand(snapshotCmd)
	faultCmd.AddCommand(addPeerCmd)
	faultCmd.AddCommand(rmPeerCmd)
	faultCmd.AddCommand(viewChangeCmd)
	faultCmd.AddCommand(backupCmd)
	faultCmd.AddCommand(restoreCmd)
	faultCmd.AddCommand(failoverCmd)
	faultCmd.AddCommand(predictCmd)
}

// NewFaultToleranceCommand returns the consolidated Cobra command tree for
// importation into the root CLI index (see README above).
func NewFaultToleranceCommand() *cobra.Command { return faultCmd }
