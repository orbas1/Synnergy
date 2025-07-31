// cmd/cli/sharding.go – Shard‑coordination CLI
// -----------------------------------------------------------------------------
// Exposes operations of the sharding subsystem under the consolidated command
// “~shard”.  All calls are proxied to the sharding daemon via a newline‑framed
// JSON‑RPC socket.
// -----------------------------------------------------------------------------
// Commands
//   leader     – get / set shard leader
//   map        – list all known shard→leader mappings
//   submit     – submit cross‑shard tx header (for manual testing)
//   pull       – pull pending receipts for our shard
//   reshard    – double the shard count (power‑of‑two upgrades)
// -----------------------------------------------------------------------------
// Environment / Config
//   SHARD_API_ADDR – host:port of sharding daemon (default "127.0.0.1:7980")
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

type shardClient struct {
	conn net.Conn
	rd   *bufio.Reader
}

func newShardClient(ctx context.Context) (*shardClient, error) {
	addr := viper.GetString("SHARD_API_ADDR")
	if addr == "" {
		addr = "127.0.0.1:7980"
	}
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to sharding daemon at %s: %w", addr, err)
	}
	return &shardClient{conn: conn, rd: bufio.NewReader(conn)}, nil
}

func (c *shardClient) Close() { _ = c.conn.Close() }

func (c *shardClient) writeJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = c.conn.Write(data)
	return err
}

func (c *shardClient) readJSON(v any) error {
	dec := json.NewDecoder(c.rd)
	return dec.Decode(v)
}

// -----------------------------------------------------------------------------
// Controller helpers
// -----------------------------------------------------------------------------

func getLeaderRPC(ctx context.Context, shard uint16) (string, error) {
	cli, err := newShardClient(ctx)
	if err != nil {
		return "", err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "get_leader", "shard": shard}); err != nil {
		return "", err
	}
	var resp struct {
		Addr  string `json:"addr"`
		Error string `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}
	return resp.Addr, nil
}

func setLeaderRPC(ctx context.Context, shard uint16, addr string) error {
	cli, err := newShardClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "set_leader", "shard": shard, "addr": addr})
}

func mapRPC(ctx context.Context) (map[string]string, error) {
	cli, err := newShardClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "map"}); err != nil {
		return nil, err
	}
	var resp struct {
		Map   map[string]string `json:"map"`
		Error string            `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return resp.Map, nil
}

func submitXSRPC(ctx context.Context, from, to uint16, hash string) error {
	cli, err := newShardClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "submit_xs", "from": from, "to": to, "hash": hash})
}

func pullRPC(ctx context.Context, shard uint16, limit int) ([]map[string]any, error) {
	cli, err := newShardClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "pull", "shard": shard, "limit": limit}); err != nil {
		return nil, err
	}
	var resp struct {
		List  []map[string]any `json:"list"`
		Error string           `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return resp.List, nil
}

func reshardRPC(ctx context.Context, bits uint8) error {
	cli, err := newShardClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "reshard", "bits": bits})
}

func rebalanceRPC(ctx context.Context, threshold float64) ([]uint16, error) {
	cli, err := newShardClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "rebalance", "threshold": threshold}); err != nil {
		return nil, err
	}
	var resp struct {
		Hot   []uint16 `json:"hot"`
		Error string   `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return resp.Hot, nil
}

// -----------------------------------------------------------------------------
// Top‑level cobra command tree
// -----------------------------------------------------------------------------

var shardCmd = &cobra.Command{
	Use:     "~shard",
	Short:   "Shard coordination & cross‑shard ops",
	Aliases: []string{"shard", "sharding"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cobra.OnInitialize(initShardConfig)
		return nil
	},
}

// leader get/set ----------------------------------------------------------------
var leaderCmd = &cobra.Command{
	Use:   "leader",
	Short: "Get or set shard leader",
}

var leaderGetCmd = &cobra.Command{
	Use:   "get [shardID]",
	Short: "Show leader for a shard",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sidU, err := strconv.ParseUint(args[0], 10, 16)
		if err != nil {
			return fmt.Errorf("invalid shardID: %w", err)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		addr, err := getLeaderRPC(ctx, uint16(sidU))
		if err != nil {
			return err
		}
		if addr == "" {
			fmt.Println("<none>")
		} else {
			fmt.Println(addr)
		}
		return nil
	},
}

var leaderSetCmd = &cobra.Command{
	Use:   "set [shardID] [addr]",
	Short: "Set leader for a shard",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sidU, err := strconv.ParseUint(args[0], 10, 16)
		if err != nil {
			return fmt.Errorf("invalid shardID: %w", err)
		}
		addr := args[1]
		if _, err := hex.DecodeString(strings.TrimPrefix(addr, "0x")); err != nil {
			return fmt.Errorf("invalid addr hex: %w", err)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		return setLeaderRPC(ctx, uint16(sidU), addr)
	},
}

// map -------------------------------------------------------------------------
var mapCmd = &cobra.Command{
	Use:   "map",
	Short: "List shard→leader mapping",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		m, err := mapRPC(ctx)
		if err != nil {
			return err
		}
		for sid, addr := range m {
			fmt.Printf("%s → %s\n", sid, addr)
		}
		return nil
	},
}

// submit ----------------------------------------------------------------------
var submitCmd = &cobra.Command{
	Use:   "submit [fromShard] [toShard] [txHash]",
	Short: "Submit cross‑shard tx header (manual)",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		fromU, err := strconv.ParseUint(args[0], 10, 16)
		if err != nil {
			return fmt.Errorf("invalid fromShard: %w", err)
		}
		toU, err := strconv.ParseUint(args[1], 10, 16)
		if err != nil {
			return fmt.Errorf("invalid toShard: %w", err)
		}
		hash := args[2]
		if _, err := hex.DecodeString(hash); err != nil || len(hash) != 64 {
			return errors.New("txHash must be 32‑byte hex")
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		return submitXSRPC(ctx, uint16(fromU), uint16(toU), hash)
	},
}

// pull ------------------------------------------------------------------------
var pullCmd = &cobra.Command{
	Use:   "pull [shardID]",
	Short: "Pull pending receipts for our shard",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		sidU, err := strconv.ParseUint(args[0], 10, 16)
		if err != nil {
			return fmt.Errorf("invalid shardID: %w", err)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		list, err := pullRPC(ctx, uint16(sidU), limit)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(list)
	},
}

// reshard ---------------------------------------------------------------------
var reshardCmd = &cobra.Command{
	Use:   "reshard [newBits]",
	Short: "Double shard count (bits > current, ≤12)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bitsU, err := strconv.ParseUint(args[0], 10, 8)
		if err != nil {
			return fmt.Errorf("invalid bits: %w", err)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()
		return reshardRPC(ctx, uint8(bitsU))
	},
}

// rebalance -------------------------------------------------------------------
var rebalanceCmd = &cobra.Command{
	Use:   "rebalance [threshold]",
	Short: "Show shards exceeding the threshold load",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		th, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid threshold: %w", err)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		list, err := rebalanceRPC(ctx, th)
		if err != nil {
			return err
		}
		for _, id := range list {
			fmt.Println(id)
		}
		return nil
	},
}

// -----------------------------------------------------------------------------
// init – config & route registration
// -----------------------------------------------------------------------------

func initShardConfig() {
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

	viper.SetDefault("SHARD_API_ADDR", "127.0.0.1:7980")
}

func init() {
	// flags
	pullCmd.Flags().Int("limit", 0, "max receipts (0=all)")

	// sub‑command tree wiring
	leaderCmd.AddCommand(leaderGetCmd)
	leaderCmd.AddCommand(leaderSetCmd)

	shardCmd.AddCommand(leaderCmd)
	shardCmd.AddCommand(mapCmd)
	shardCmd.AddCommand(submitCmd)
	shardCmd.AddCommand(pullCmd)
	shardCmd.AddCommand(reshardCmd)
	shardCmd.AddCommand(rebalanceCmd)
}

// NewShardingCommand exposes the consolidated command tree.
func NewShardingCommand() *cobra.Command { return shardCmd }
