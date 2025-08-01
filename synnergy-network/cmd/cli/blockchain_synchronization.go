package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// -----------------------------------------------------------
// Middleware – framed JSON/TCP client
// -----------------------------------------------------------

type syncClient struct {
	conn net.Conn
	rd   *bufio.Reader
}

func newSyncClient(ctx context.Context) (*syncClient, error) {
	addr := viper.GetString("SYNC_API_ADDR")
	if addr == "" {
		addr = "127.0.0.1:7960"
	}
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to sync daemon at %s: %w", addr, err)
	}
	return &syncClient{conn: conn, rd: bufio.NewReader(conn)}, nil
}

func (c *syncClient) Close() { _ = c.conn.Close() }

func (c *syncClient) writeJSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = c.conn.Write(b)
	return err
}

func (c *syncClient) readJSON(v any) error {
	dec := json.NewDecoder(c.rd)
	return dec.Decode(v)
}

// -----------------------------------------------------------
// RPC helpers
// -----------------------------------------------------------

func syncStartRPC(ctx context.Context) error {
	cli, err := newSyncClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "start"})
}

func syncStopRPC(ctx context.Context) error {
	cli, err := newSyncClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "stop"})
}

func syncStatusRPC(ctx context.Context) (map[string]any, error) {
	cli, err := newSyncClient(ctx)
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

func syncOnceRPC(ctx context.Context) error {
	cli, err := newSyncClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "once"})
}

// -----------------------------------------------------------
// Cobra commands
// -----------------------------------------------------------

var syncCmd = &cobra.Command{
	Use:     "~sync",
	Short:   "Blockchain synchronization manager",
	Aliases: []string{"sync", "synchronization"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cobra.OnInitialize(initSyncConfig)
		return nil
	},
}

var syncStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the sync manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		return syncStartRPC(ctx)
	},
}

var syncStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the sync manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		return syncStopRPC(ctx)
	},
}

var syncStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync status",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		data, err := syncStatusRPC(ctx)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	},
}

var syncOnceCmd = &cobra.Command{
	Use:   "once",
	Short: "Perform a single sync round",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer cancel()
		return syncOnceRPC(ctx)
	},
}

func initSyncConfig() {
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
	viper.SetDefault("SYNC_API_ADDR", "127.0.0.1:7960")
}

func init() {
	syncCmd.AddCommand(syncStartCmd)
	syncCmd.AddCommand(syncStopCmd)
	syncCmd.AddCommand(syncStatusCmd)
	syncCmd.AddCommand(syncOnceCmd)
}

// NewSyncCommand exposes the root command for ~sync.
func NewSyncCommand() *cobra.Command { return syncCmd }
