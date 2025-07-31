// cmd/cli/rollups.go – Optimistic‑roll‑up CLI
// -----------------------------------------------------------------------------
// Consolidated under the route “~rollup”.  Exposes administrative and operator
// commands for submitting batches, challenging fraud, finalising, and querying
// roll‑up status.
// -----------------------------------------------------------------------------
// Commands
//   • submit      – create a batch from a list of tx hashes (hex‑32) or stdin
//   • challenge   – lodge a fraud proof against a batch
//   • finalize    – finalise or revert a batch after the challenge window
//   • info        – show batch header & state
//   • list        – list recent batches (paginated)
// -----------------------------------------------------------------------------
// Environment / Config
//   ROLLUP_API_ADDR – host:port of roll‑up daemon (default "127.0.0.1:7960")
// -----------------------------------------------------------------------------

package cli

import (
	"bufio"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core"
)

// -----------------------------------------------------------------------------
// Middleware – framed JSON/TCP client
// -----------------------------------------------------------------------------

type rollClient struct {
	conn net.Conn
	rd   *bufio.Reader
}

func newRollClient(ctx context.Context) (*rollClient, error) {
	addr := viper.GetString("ROLLUP_API_ADDR")
	if addr == "" {
		addr = "127.0.0.1:7960"
	}
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to roll‑up daemon at %s: %w", addr, err)
	}
	return &rollClient{conn: conn, rd: bufio.NewReader(conn)}, nil
}

func (c *rollClient) Close() { _ = c.conn.Close() }

func (c *rollClient) writeJSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = c.conn.Write(b)
	return err
}

func (c *rollClient) readJSON(v any) error {
	dec := json.NewDecoder(c.rd)
	return dec.Decode(v)
}

// -----------------------------------------------------------------------------
// Controller helpers
// -----------------------------------------------------------------------------

func submitBatchRPC(ctx context.Context, txHashes [][]byte, preRoot string, submitter string) (uint64, error) {
	cli, err := newRollClient(ctx)
	if err != nil {
		return 0, err
	}
	defer cli.Close()
	return sendAndGetID(cli, map[string]any{
		"action":    "submit",
		"tx_hashes": txHashes,
		"pre_root":  preRoot,
		"submitter": submitter,
	})
}

func sendAndGetID(cli *rollClient, payload map[string]any) (uint64, error) {
	if err := cli.writeJSON(payload); err != nil {
		return 0, err
	}
	var resp struct {
		ID    uint64 `json:"id"`
		Error string `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return 0, err
	}
	if resp.Error != "" {
		return 0, errors.New(resp.Error)
	}
	return resp.ID, nil
}

func challengeRPC(ctx context.Context, batchID uint64, txIdx uint32, proof [][]byte) error {
	cli, err := newRollClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{
		"action":   "challenge",
		"batch_id": batchID,
		"tx_idx":   txIdx,
		"proof":    proof,
	})
}

func finalizeRPC(ctx context.Context, batchID uint64) error {
	cli, err := newRollClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "finalize", "batch_id": batchID})
}

func infoRPC(ctx context.Context, batchID uint64) (*core.BatchHeader, core.BatchState, error) {
	cli, err := newRollClient(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "info", "batch_id": batchID}); err != nil {
		return nil, 0, err
	}
	var resp struct {
		Header core.BatchHeader `json:"header"`
		State  core.BatchState  `json:"state"`
		Error  string           `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return nil, 0, err
	}
	if resp.Error != "" {
		return nil, 0, errors.New(resp.Error)
	}
	return &resp.Header, resp.State, nil
}

func listRPC(ctx context.Context, limit int) ([]struct {
	Header core.BatchHeader `json:"header"`
	State  core.BatchState  `json:"state"`
}, error) {
	cli, err := newRollClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "list", "limit": limit}); err != nil {
		return nil, err
	}
	var resp struct {
		List []struct {
			Header core.BatchHeader `json:"header"`
			State  core.BatchState  `json:"state"`
		} `json:"list"`
		Error string `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return resp.List, nil
}

// -----------------------------------------------------------------------------
// Top‑level Cobra commands
// -----------------------------------------------------------------------------

var rollCmd = &cobra.Command{
	Use:     "~rollup",
	Short:   "Optimistic roll‑up operations",
	Aliases: []string{"rollup", "rlp"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cobra.OnInitialize(initRollConfig)
		return nil
	},
}

// submit ----------------------------------------------------------------------
var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a new batch (reads tx hashes from flags or stdin)",
	RunE: func(cmd *cobra.Command, args []string) error {
		preRoot, _ := cmd.Flags().GetString("pre-root")
		submitter, _ := cmd.Flags().GetString("submitter")
		txFlag, _ := cmd.Flags().GetString("txs")

		var hashes [][]byte
		if txFlag != "" {
			for _, h := range strings.Split(txFlag, ",") {
				bh, err := hex.DecodeString(strings.TrimSpace(h))
				if err != nil || len(bh) != 32 {
					return fmt.Errorf("invalid tx hash %q", h)
				}
				hashes = append(hashes, bh)
			}
		} else {
			// read newline‑separated hex hashes from stdin
			stdinHashes, err := readHashes(os.Stdin)
			if err != nil {
				return err
			}
			hashes = stdinHashes
		}
		if len(hashes) == 0 {
			return errors.New("no tx hashes provided")
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 4*time.Second)
		defer cancel()
		id, err := submitBatchRPC(ctx, hashes, preRoot, submitter)
		if err != nil {
			return err
		}
		fmt.Printf("batch submitted: %d\n", id)
		return nil
	},
}

func readHashes(r io.Reader) ([][]byte, error) {
	var res [][]byte
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		bh, err := hex.DecodeString(line)
		if err != nil || len(bh) != 32 {
			return nil, fmt.Errorf("invalid tx hash %q", line)
		}
		res = append(res, bh)
	}
	return res, sc.Err()
}

// challenge -------------------------------------------------------------------
var challengeCmd = &cobra.Command{
	Use:   "challenge [batchID] [txIdx] [proof‑hex…]",
	Short: "Submit a fraud proof for a batch",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid batchID: %w", err)
		}
		idxU, err := strconv.ParseUint(args[1], 10, 32)
		if err != nil {
			return fmt.Errorf("invalid txIdx: %w", err)
		}
		var proof [][]byte
		for _, p := range args[2:] {
			b, err := hex.DecodeString(p)
			if err != nil {
				return fmt.Errorf("invalid proof chunk %q", p)
			}
			proof = append(proof, b)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		return challengeRPC(ctx, id, uint32(idxU), proof)
	},
}

// finalize --------------------------------------------------------------------
var finalizeCmd = &cobra.Command{
	Use:   "finalize [batchID]",
	Short: "Finalize (or revert) a batch after challenge period",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid batchID: %w", err)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		return finalizeRPC(ctx, id)
	},
}

// info ------------------------------------------------------------------------
var infoCmd = &cobra.Command{
	Use:   "info [batchID]",
	Short: "Display batch header & state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid batchID: %w", err)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		hdr, state, err := infoRPC(ctx, id)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(struct {
			Header core.BatchHeader `json:"header"`
			State  core.BatchState  `json:"state"`
		}{*hdr, state})
	},
}

// list ------------------------------------------------------------------------
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent batches",
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		format := viper.GetString("output.format")
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		list, err := listRPC(ctx, limit)
		if err != nil {
			return err
		}
		switch format {
		case "json":
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(list)
		default:
			fmt.Printf("%-5s %-12s %-8s %-10s\n", "ID", "Timestamp", "TXs", "State")
			for _, e := range list {
				fmt.Printf("%-5d %-12d %-8d %-10s\n", e.Header.BatchID, e.Header.Timestamp, e.Header.TXCount, e.State)
			}
			return nil
		}
	},
}

// -----------------------------------------------------------------------------
// init – config & route wiring
// -----------------------------------------------------------------------------

func initRollConfig() {
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

	viper.SetDefault("ROLLUP_API_ADDR", "127.0.0.1:7960")
	viper.SetDefault("output.format", "table")
}

func init() {
	submitCmd.Flags().String("pre-root", "", "pre‑state root hex (optional)")
	submitCmd.Flags().String("txs", "", "comma‑separated list of tx hashes (hex32); omit to read from stdin")
	submitCmd.Flags().String("submitter", "", "submitter address hex")

	listCmd.Flags().Int("limit", 10, "max batches to list")
	listCmd.Flags().StringP("format", "f", "table", "output format: table|json")
	_ = viper.BindPFlag("output.format", listCmd.Flags().Lookup("format"))

	// Register sub‑commands
	rollCmd.AddCommand(submitCmd)
	rollCmd.AddCommand(challengeCmd)
	rollCmd.AddCommand(finalizeCmd)
	rollCmd.AddCommand(infoCmd)
	rollCmd.AddCommand(listCmd)
}

// NewRollupCommand exposes the consolidated command tree.
func NewRollupCommand() *cobra.Command { return rollCmd }
