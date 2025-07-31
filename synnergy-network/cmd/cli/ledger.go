// cmd/cli/ledger.go – Ledger inspection & maintenance CLI
// -----------------------------------------------------------------------------
// Consolidated under route “~ledger”. Provides read‑only chain inspection as
// well as administrative token operations (mint/transfer) via JSON‑over‑TCP to
// the ledger daemon. All sub‑commands are declared first for readability, with
// route wiring at the bottom and `NewLedgerCommand()` exported for importation
// into the root CLI index.
// -----------------------------------------------------------------------------
// Examples
//   synnergy ~ledger head                          # height + last block hash
//   synnergy ~ledger block 123 --format=json       # inspect block
//   synnergy ~ledger balance 0xabc…                # token balances
//   synnergy ~ledger utxo 0xabc… --limit=20
//   synnergy ~ledger pool --limit=10 --format=json # mem‑pool slice
//   synnergy ~ledger mint 0xabc… --token=SYNR --amount=1000
//   synnergy ~ledger transfer 0xabc… 0xdef… --token=SYNR --amount=250
// -----------------------------------------------------------------------------
// Environment
//   LEDGER_API_ADDR – host:port of ledger daemon (default "127.0.0.1:7900")
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
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core"
)

// -----------------------------------------------------------------------------
// Middleware – framed JSON/TCP client
// -----------------------------------------------------------------------------

type ledgerClient struct {
	conn net.Conn
	r    *bufio.Reader
}

func newLedgerClient(ctx context.Context) (*ledgerClient, error) {
	addr := viper.GetString("LEDGER_API_ADDR")
	if addr == "" {
		addr = "127.0.0.1:7900"
	}
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to ledger daemon at %s: %w", addr, err)
	}
	return &ledgerClient{conn: conn, r: bufio.NewReader(conn)}, nil
}

func (c *ledgerClient) Close() { _ = c.conn.Close() }

func (c *ledgerClient) writeJSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = c.conn.Write(b)
	return err
}

func (c *ledgerClient) readJSON(v any) error {
	dec := json.NewDecoder(c.r)
	return dec.Decode(v)
}

// -----------------------------------------------------------------------------
// Controller helpers
// -----------------------------------------------------------------------------

func headRPC(ctx context.Context) (height uint64, hash string, err error) {
	cli, errN := newLedgerClient(ctx)
	if errN != nil {
		err = errN
		return
	}
	defer cli.Close()
	if err = cli.writeJSON(map[string]any{"action": "head"}); err != nil {
		return
	}
	var resp struct {
		Height uint64 `json:"height"`
		Hash   string `json:"hash"`
		Error  string `json:"error,omitempty"`
	}
	if err = cli.readJSON(&resp); err != nil {
		return
	}
	if resp.Error != "" {
		err = errors.New(resp.Error)
		return
	}
	height, hash = resp.Height, resp.Hash
	return
}

func blockRPC(ctx context.Context, h uint64) (*core.Block, error) {
	cli, err := newLedgerClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "block", "height": h}); err != nil {
		return nil, err
	}
	var resp struct {
		Block core.Block `json:"block"`
		Error string     `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return &resp.Block, nil
}

func balanceRPC(ctx context.Context, addr string) (map[string]uint64, error) {
	cli, err := newLedgerClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "balance", "addr": addr}); err != nil {
		return nil, err
	}
	var resp struct {
		Bal   map[string]uint64 `json:"bal"`
		Error string            `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return resp.Bal, nil
}

func utxoRPC(ctx context.Context, addr string, limit int) ([]core.UTXO, error) {
	cli, err := newLedgerClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "utxo", "addr": addr, "limit": limit}); err != nil {
		return nil, err
	}
	var resp struct {
		List  []core.UTXO `json:"list"`
		Error string      `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return resp.List, nil
}

func poolRPC(ctx context.Context, limit int) ([]core.Transaction, error) {
	cli, err := newLedgerClient(ctx)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "pool", "limit": limit}); err != nil {
		return nil, err
	}
	var resp struct {
		List  []core.Transaction `json:"list"`
		Error string             `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return resp.List, nil
}

func mintRPC(ctx context.Context, addr, token string, amt uint64) error {
	cli, err := newLedgerClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "mint", "addr": addr, "token": token, "amount": amt})
}

func transferRPC(ctx context.Context, from, to, token string, amt uint64) error {
	cli, err := newLedgerClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()
	return cli.writeJSON(map[string]any{"action": "transfer", "from": from, "to": to, "token": token, "amount": amt})
}

// -----------------------------------------------------------------------------
// Top-level Cobra commands
// -----------------------------------------------------------------------------

var ledgerCmd = &cobra.Command{
	Use:     "~ledger",
	Short:   "Ledger inspection & maintenance",
	Aliases: []string{"ledger", "ldg"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cobra.OnInitialize(initLedgerConfig)
		return nil
	},
}

// head ------------------------------------------------------------------------
var headCmd = &cobra.Command{
	Use:   "head",
	Short: "Show chain height and latest block hash",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		h, hash, err := headRPC(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("height: %d\nhash:   %s\n", h, hash)
		return nil
	},
}

// block -----------------------------------------------------------------------
var blockCmd = &cobra.Command{
	Use:   "block [height]",
	Short: "Fetch block by height",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fStr, _ := cmd.Flags().GetString("format")
		heightU, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid height: %w", err)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		blk, err := blockRPC(ctx, heightU)
		if err != nil {
			return err
		}
		switch fStr {
		case "json":
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(blk)
		default:
			fmt.Printf("Block %d: %d tx, prev=%x\n", blk.Header.Height, len(blk.Transactions), blk.Header.PrevHash)
			return nil
		}
	},
}

// balance ---------------------------------------------------------------------
var balanceCmd = &cobra.Command{
	Use:   "balance [addr]",
	Short: "Show token balances of an address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		bal, err := balanceRPC(ctx, args[0])
		if err != nil {
			return err
		}
		if len(bal) == 0 {
			fmt.Println("no balance")
			return nil
		}
		for t, amt := range bal {
			fmt.Printf("%s: %d\n", t, amt)
		}
		return nil
	},
}

// utxo ------------------------------------------------------------------------
var utxoCmd = &cobra.Command{
	Use:   "utxo [addr]",
	Short: "List UTXOs for address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		list, err := utxoRPC(ctx, args[0], limit)
		if err != nil {
			return err
		}
		for _, u := range list {
			txid := hex.EncodeToString(u.TxID[:])
			fmt.Printf("%s:%d  value=%d\n", txid, u.Index, u.Output.Value)
		}
		return nil
	},
}

// pool ------------------------------------------------------------------------
var ledgerPoolCmd = &cobra.Command{
	Use:   "pool",
	Short: "List pending transactions in mem‑pool",
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		format := viper.GetString("output.format")
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		list, err := poolRPC(ctx, limit)
		if err != nil {
			return err
		}
		switch format {
		case "json":
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(list)
		default:
			for _, tx := range list {
				fmt.Printf("%x  size=%d bytes\n", tx.ID(), len(tx.Payload))
			}
			return nil
		}
	},
}

// mint ------------------------------------------------------------------------
var mintCmd = &cobra.Command{
	Use:   "mint [addr]",
	Short: "Mint tokens to an address (admin)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, _ := cmd.Flags().GetString("token")
		amtStr, _ := cmd.Flags().GetString("amount")
		if token == "" || amtStr == "" {
			return errors.New("--token and --amount required")
		}
		amt, err := strconv.ParseUint(amtStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid --amount: %w", err)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		return mintRPC(ctx, args[0], token, amt)
	},
}

// transfer --------------------------------------------------------------------
var transferCmd = &cobra.Command{
	Use:   "transfer [from] [to]",
	Short: "Transfer tokens between addresses",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, _ := cmd.Flags().GetString("token")
		amtStr, _ := cmd.Flags().GetString("amount")
		if token == "" || amtStr == "" {
			return errors.New("--token and --amount required")
		}
		amt, err := strconv.ParseUint(amtStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid --amount: %w", err)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		return transferRPC(ctx, args[0], args[1], token, amt)
	},
}

// -----------------------------------------------------------------------------
// init – config + route wiring
// -----------------------------------------------------------------------------

func initLedgerConfig() {
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

	viper.SetDefault("LEDGER_API_ADDR", "127.0.0.1:7900")
	viper.SetDefault("output.format", "table")
}

func init() {
	// flags
	blockCmd.Flags().StringP("format", "f", "table", "output format: table|json")
	_ = viper.BindPFlag("output.format", blockCmd.Flags().Lookup("format"))

	utxoCmd.Flags().Int("limit", 0, "max entries (0=all)")
	ledgerPoolCmd.Flags().Int("limit", 0, "max transactions (0=all)")
	ledgerPoolCmd.Flags().StringP("format", "f", "table", "output format: table|json")
	_ = viper.BindPFlag("output.format", ledgerPoolCmd.Flags().Lookup("format"))

	mintCmd.Flags().String("token", "", "token symbol or ID")
	mintCmd.Flags().String("amount", "", "amount to mint")

	transferCmd.Flags().String("token", "", "token symbol or ID")
	transferCmd.Flags().String("amount", "", "amount to transfer")

	// wire routes
	ledgerCmd.AddCommand(headCmd)
	ledgerCmd.AddCommand(blockCmd)
	ledgerCmd.AddCommand(balanceCmd)
	ledgerCmd.AddCommand(utxoCmd)
	ledgerCmd.AddCommand(ledgerPoolCmd)
	ledgerCmd.AddCommand(mintCmd)
	ledgerCmd.AddCommand(transferCmd)
}

// NewLedgerCommand exposes the consolidated command tree.
func NewLedgerCommand() *cobra.Command { return ledgerCmd }
