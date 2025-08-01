// cmd/cli/energy_efficiency.go - Energy efficiency CLI
// -----------------------------------------------------------------------------
// Provides commands under the route "energy" for recording validator transaction
// statistics and viewing efficiency metrics.
// -----------------------------------------------------------------------------

package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// -----------------------------------------------------------------------------
// Middleware: simple JSON framed TCP client
// -----------------------------------------------------------------------------

type effClient struct {
	conn net.Conn
	r    *bufio.Reader
}

func newEffClient(ctx context.Context) (*effClient, error) {
	addr := viper.GetString("ENERGY_API_ADDR")
	if addr == "" {
		addr = "127.0.0.1:7810"
	}
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to energy daemon at %s: %w", addr, err)
	}
	return &effClient{conn: conn, r: bufio.NewReader(conn)}, nil
}

func (c *effClient) Close() { _ = c.conn.Close() }

func (c *effClient) writeJSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = c.conn.Write(b)
	return err
}

func (c *effClient) readJSON(v any) error {
	dec := json.NewDecoder(c.r)
	return dec.Decode(v)
}

// -----------------------------------------------------------------------------
// Controller helpers
// -----------------------------------------------------------------------------

func recordStatsRPC(ctx context.Context, validator string, txs uint64, kwh float64) error {
	cli, err := newEffClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{
		"action":    "record",
		"validator": validator,
		"txs":       txs,
		"energy":    kwh,
	})
}

func efficiencyOfRPC(ctx context.Context, validator string) (float64, error) {
	cli, err := newEffClient(ctx)
	if err != nil {
		return 0, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "efficiency", "validator": validator}); err != nil {
		return 0, err
	}
	var resp struct{ Score float64 }
	if err := cli.readJSON(&resp); err != nil {
		return 0, err
	}
	return resp.Score, nil
}

func networkAvgRPC(ctx context.Context) (float64, error) {
	cli, err := newEffClient(ctx)
	if err != nil {
		return 0, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "network"}); err != nil {
		return 0, err
	}
	var resp struct{ Score float64 }
	if err := cli.readJSON(&resp); err != nil {
		return 0, err
	}
	return resp.Score, nil
}

// -----------------------------------------------------------------------------
// Cobra commands
// -----------------------------------------------------------------------------

var energyCmd = &cobra.Command{Use: "energy", Short: "Energy efficiency tools"}

var energyRecordCmd = &cobra.Command{
	Use:   "record [validator-addr]",
	Short: "Record processed transactions and energy use",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		txsStr, _ := cmd.Flags().GetString("txs")
		energyStr, _ := cmd.Flags().GetString("energy")
		txs, err := strconv.ParseUint(txsStr, 10, 64)
		if err != nil {
			return err
		}
		kwh, err := strconv.ParseFloat(energyStr, 64)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		return recordStatsRPC(ctx, args[0], txs, kwh)
	},
}

var energyEffCmd = &cobra.Command{
	Use:   "efficiency [validator-addr]",
	Short: "Show transactions per kWh for a validator",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		score, err := efficiencyOfRPC(ctx, args[0])
		if err != nil {
			return err
		}
		fmt.Printf("%.2f tx/kWh\n", score)
		return nil
	},
}

var energyNetCmd = &cobra.Command{
	Use:   "network",
	Short: "Show network average transactions per kWh",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		score, err := networkAvgRPC(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("%.2f tx/kWh\n", score)
		return nil
	},
}

func init() {
	energyRecordCmd.Flags().String("txs", "", "number of transactions")
	energyRecordCmd.Flags().String("energy", "", "energy used in kWh")

	energyCmd.AddCommand(energyRecordCmd)
	energyCmd.AddCommand(energyEffCmd)
	energyCmd.AddCommand(energyNetCmd)
}

// EnergyCmd is the exported command tree.
var EnergyCmd = energyCmd
