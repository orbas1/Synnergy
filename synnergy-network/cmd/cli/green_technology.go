// cmd/cli/green_technology.go – Sustainability & carbon‑accounting CLI
// -----------------------------------------------------------------------------
// This command tree is mounted under the consolidated route “~green”.  It lets
// operators record validator energy usage / carbon emissions, retire carbon‑
// offset credits, force a certification run for the current epoch, and inspect
// certificates or throttle status.
//
// • All top‑level Cobra *commands* are declared first (for readability).
// • Middleware provides a small JSON‑over‑TCP client with newline framing.
// • Controller helpers wrap the low‑level RPC calls.
// • An `init()` block wires flags and sub‑commands; `NewGreenCommand()` returns
//   the consolidated tree for `rootCmd.AddCommand()`.
// -----------------------------------------------------------------------------
// Example usage
//   synnergy ~green usage 76c2…ffae --energy=1234.5 --carbon=321.0
//   synnergy ~green offset 76c2…ffae --kg=500
//   synnergy ~green certify           # recompute certificates for all nodes
//   synnergy ~green cert 76c2…ffae    # print node certificate
//   synnergy ~green list --format=json
// -----------------------------------------------------------------------------
// Environment
//   GREEN_API_ADDR – host:port of the green‑technology daemon (default
//                    "127.0.0.1:7800")
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
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core"
)

// -----------------------------------------------------------------------------
// Middleware – thin framed‑JSON TCP client
// -----------------------------------------------------------------------------

type greenClient struct {
	conn net.Conn
	r    *bufio.Reader
}

func newGreenClient(ctx context.Context) (*greenClient, error) {
	addr := viper.GetString("GREEN_API_ADDR")
	if addr == "" {
		addr = "127.0.0.1:7800"
	}
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to green‑tech daemon at %s: %w", addr, err)
	}
	return &greenClient{conn: conn, r: bufio.NewReader(conn)}, nil
}

func (c *greenClient) Close() { _ = c.conn.Close() }

func (c *greenClient) writeJSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = c.conn.Write(b)
	return err
}

func (c *greenClient) readJSON(v any) error {
	dec := json.NewDecoder(c.r)
	return dec.Decode(v)
}

// -----------------------------------------------------------------------------
// Controller helpers – thin wrappers around RPC
// -----------------------------------------------------------------------------

func recordUsageRPC(ctx context.Context, validator string, energy, carbon float64) error {
	cli, err := newGreenClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{
		"action":    "record_usage",
		"validator": validator,
		"energy":    energy,
		"carbon":    carbon,
	})
}

func recordOffsetRPC(ctx context.Context, validator string, kg float64) error {
	cli, err := newGreenClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{
		"action":    "record_offset",
		"validator": validator,
		"kg":        kg,
	})
}

func certifyRPC(ctx context.Context) error {
	cli, err := newGreenClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "certify"})
}

func certificateOfRPC(ctx context.Context, addr string) (core.Certificate, error) {
	cli, err := newGreenClient(ctx)
	if err != nil {
		return "", err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "certificate_of", "addr": addr}); err != nil {
		return "", err
	}
	var resp struct {
		Cert  core.Certificate `json:"cert"`
		Error string           `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}
	return resp.Cert, nil
}

func shouldThrottleRPC(ctx context.Context, addr string) (bool, error) {
	cli, err := newGreenClient(ctx)
	if err != nil {
		return false, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "should_throttle", "addr": addr}); err != nil {
		return false, err
	}
	var resp struct {
		Throttle bool   `json:"throttle"`
		Error    string `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return false, err
	}
	if resp.Error != "" {
		return false, errors.New(resp.Error)
	}
	return resp.Throttle, nil
}

func listCertsRPC(ctx context.Context) ([]struct {
	Address string           `json:"address"`
	Cert    core.Certificate `json:"cert"`
	Score   float64          `json:"score"`
}, error) {
	cli, err := newGreenClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "list"}); err != nil {
		return nil, err
	}
	var resp struct {
		Certs []struct {
			Address string           `json:"address"`
			Cert    core.Certificate `json:"cert"`
			Score   float64          `json:"score"`
		} `json:"certs"`
		Error string `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return resp.Certs, nil
}

// -----------------------------------------------------------------------------
// Top‑level Cobra commands
// -----------------------------------------------------------------------------

var greenCmd = &cobra.Command{
	Use:     "~green",
	Short:   "Validator sustainability & carbon‑offset operations",
	Aliases: []string{"green", "eco"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cobra.OnInitialize(initGreenConfig)
		return nil
	},
}

// usage -----------------------------------------------------------------------
var usageCmd = &cobra.Command{
	Use:   "usage [validator‑addr]",
	Short: "Record energy & carbon usage for a validator",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		energyStr, _ := cmd.Flags().GetString("energy")
		carbonStr, _ := cmd.Flags().GetString("carbon")
		if energyStr == "" || carbonStr == "" {
			return errors.New("--energy and --carbon required")
		}
		energy, err := strconv.ParseFloat(energyStr, 64)
		if err != nil {
			return fmt.Errorf("invalid --energy: %w", err)
		}
		carbon, err := strconv.ParseFloat(carbonStr, 64)
		if err != nil {
			return fmt.Errorf("invalid --carbon: %w", err)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		return recordUsageRPC(ctx, args[0], energy, carbon)
	},
}

// offset ----------------------------------------------------------------------
var offsetCmd = &cobra.Command{
	Use:   "offset [validator‑addr]",
	Short: "Record carbon offset credits (kg CO₂e)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kgStr, _ := cmd.Flags().GetString("kg")
		if kgStr == "" {
			return errors.New("--kg required")
		}
		kg, err := strconv.ParseFloat(kgStr, 64)
		if err != nil {
			return fmt.Errorf("invalid --kg: %w", err)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		return recordOffsetRPC(ctx, args[0], kg)
	},
}

// certify ---------------------------------------------------------------------
var certifyCmd = &cobra.Command{
	Use:   "certify",
	Short: "Run certificate recomputation now",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		return certifyRPC(ctx)
	},
}

// cert ------------------------------------------------------------------------
var certCmd = &cobra.Command{
	Use:   "cert [validator‑addr]",
	Short: "Show certificate of a validator",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		cert, err := certificateOfRPC(ctx, args[0])
		if err != nil {
			return err
		}
		fmt.Println(cert)
		return nil
	},
}

// throttle --------------------------------------------------------------------
var throttleCmd = &cobra.Command{
	Use:   "throttle [validator‑addr]",
	Short: "Check if validator should be throttled",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		yes, err := shouldThrottleRPC(ctx, args[0])
		if err != nil {
			return err
		}
		if yes {
			fmt.Println("yes")
		} else {
			fmt.Println("no")
		}
		return nil
	},
}

// list ------------------------------------------------------------------------
var greenListCmd = &cobra.Command{
	Use:   "list",
	Short: "List certificates for all validators",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 4*time.Second)
		defer cancel()
		certs, err := listCertsRPC(ctx)
		if err != nil {
			return err
		}
		switch viper.GetString("output.format") {
		case "json":
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(certs)
		default:
			fmt.Printf("%-24s %-6s %s\n", "Validator", "Score", "Cert")
			for _, c := range certs {
				fmt.Printf("%-24s %6.2f %s\n", c.Address, c.Score, c.Cert)
			}
			return nil
		}
	},
}

// -----------------------------------------------------------------------------
// init – config & route wiring
// -----------------------------------------------------------------------------

func initGreenConfig() {
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

	viper.SetDefault("GREEN_API_ADDR", "127.0.0.1:7800")
	viper.SetDefault("output.format", "table")
}

func init() {
	usageCmd.Flags().String("energy", "", "energy consumption in kWh")
	usageCmd.Flags().String("carbon", "", "CO₂ emissions in kg")

	offsetCmd.Flags().String("kg", "", "offset amount in kg CO₂e")

	greenListCmd.Flags().StringP("format", "f", "table", "output format: table|json")
	_ = viper.BindPFlag("output.format", greenListCmd.Flags().Lookup("format"))

	// Register sub‑commands
	greenCmd.AddCommand(usageCmd)
	greenCmd.AddCommand(offsetCmd)
	greenCmd.AddCommand(certifyCmd)
	greenCmd.AddCommand(certCmd)
	greenCmd.AddCommand(throttleCmd)
	greenCmd.AddCommand(greenListCmd)
}

// NewGreenCommand exposes the consolidated green‑tech command tree.
func NewGreenCommand() *cobra.Command { return greenCmd }
