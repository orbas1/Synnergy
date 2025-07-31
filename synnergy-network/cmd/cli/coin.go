package cli

// -----------------------------------------------------------------------------
// coin.go – Synthron CLI middleware for the native SYNN asset
// -----------------------------------------------------------------------------
// Commands exposed after `RegisterCoin(rootCmd)`:
//   ~coin ~mint    <address> <amount>
//   ~coin ~supply
//   ~coin ~balance <address>
//
// Every symbol is prefixed **coin*** to avoid clashes with other middleware
// files (tokens.go, contracts.go, etc.).
// -----------------------------------------------------------------------------

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"synnergy-network/core"
)

// -----------------------------------------------------------------------------
// Globals – initialised once in coinInitMiddleware
// -----------------------------------------------------------------------------

var (
	coinLedger *core.Ledger
	coinMgr    *core.Coin
	coinOnce   sync.Once
)

// -----------------------------------------------------------------------------
// Middleware – load env, create ledger & coin manager
// -----------------------------------------------------------------------------

func coinInitMiddleware(cmd *cobra.Command, _ []string) error {
	var err error
	coinOnce.Do(func() {
		// 1) .env → ENV
		_ = godotenv.Load()

		// 2) Logging level
		lvl := os.Getenv("LOG_LEVEL")
		if lvl == "" {
			lvl = "info"
		}
		lv, e := logrus.ParseLevel(lvl)
		if e != nil {
			err = e
			return
		}
		logrus.SetLevel(lv)

		// 3) Build ledger config (WAL + snapshot paths configurable via ENV)
		walPath := coinEnvOr("LEDGER_WAL", "./ledger.wal")
		snapPath := coinEnvOr("LEDGER_SNAPSHOT", "./ledger.snap")
		interval := coinEnvOrInt("LEDGER_SNAPSHOT_INTERVAL", 100)

		coinLedger, e = core.NewLedger(core.LedgerConfig{
			WALPath:          walPath,
			SnapshotPath:     snapPath,
			SnapshotInterval: interval,
		})
		if e != nil {
			err = e
			return
		}

		coinMgr, e = core.NewCoin(coinLedger)
		if e != nil {
			err = e
		}
	})
	return err
}

func coinEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
func coinEnvOrInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// -----------------------------------------------------------------------------
// Helper parsing utils
// -----------------------------------------------------------------------------

func coinDecodeAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("invalid address")
	}
	copy(a[:], b)
	return a, nil
}

func coinParseAmt(s string) (uint64, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil || n == 0 {
		return 0, fmt.Errorf("amount must be positive uint64")
	}
	return n, nil
}

// -----------------------------------------------------------------------------
// Controllers
// -----------------------------------------------------------------------------

func coinHandleMint(cmd *cobra.Command, args []string) error {
	addr, err := coinDecodeAddr(args[0])
	if err != nil {
		return err
	}
	amt, err := coinParseAmt(args[1])
	if err != nil {
		return err
	}
	if err := coinMgr.Mint(addr[:], amt); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "minted %d %s to %s\n", amt, core.Code, args[0])
	return nil
}

func coinHandleSupply(cmd *cobra.Command, _ []string) error {
	fmt.Fprintf(cmd.OutOrStdout(), "%d\n", coinMgr.TotalSupply())
	return nil
}

func coinHandleBalance(cmd *cobra.Command, args []string) error {
	addr, err := coinDecodeAddr(args[0])
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%d\n", coinMgr.BalanceOf(addr[:]))
	return nil
}

func coinHandleTransfer(cmd *cobra.Command, args []string) error {
	from, err := coinDecodeAddr(args[0])
	if err != nil {
		return err
	}
	to, err := coinDecodeAddr(args[1])
	if err != nil {
		return err
	}
	amt, err := coinParseAmt(args[2])
	if err != nil {
		return err
	}
	if err := coinMgr.Transfer(from[:], to[:], amt); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "transferred %d %s from %s to %s\n", amt, core.Code, args[0], args[1])
	return nil
}

func coinHandleBurn(cmd *cobra.Command, args []string) error {
	addr, err := coinDecodeAddr(args[0])
	if err != nil {
		return err
	}
	amt, err := coinParseAmt(args[1])
	if err != nil {
		return err
	}
	if err := coinMgr.Burn(addr[:], amt); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "burned %d %s from %s\n", amt, core.Code, args[0])
	return nil
}

// -----------------------------------------------------------------------------
// Cobra command tree
// -----------------------------------------------------------------------------

var coinRootCmd = &cobra.Command{
	Use:               "coin",
	Short:             "Native SYNN coin operations",
	PersistentPreRunE: coinInitMiddleware,
}

var coinMintCmd = &cobra.Command{Use: "mint <addr> <amt>", Short: "Mint SYNN", Args: cobra.ExactArgs(2), RunE: coinHandleMint}
var coinSupplyCmd = &cobra.Command{Use: "supply", Short: "Total supply", Args: cobra.NoArgs, RunE: coinHandleSupply}
var coinBalCmd = &cobra.Command{Use: "balance <addr>", Short: "Balance", Args: cobra.ExactArgs(1), RunE: coinHandleBalance}
var coinTransferCmd = &cobra.Command{Use: "transfer <from> <to> <amt>", Short: "Transfer SYNN", Args: cobra.ExactArgs(3), RunE: coinHandleTransfer}
var coinBurnCmd = &cobra.Command{Use: "burn <addr> <amt>", Short: "Burn SYNN", Args: cobra.ExactArgs(2), RunE: coinHandleBurn}

func init() {
	coinRootCmd.AddCommand(coinMintCmd, coinSupplyCmd, coinBalCmd, coinTransferCmd, coinBurnCmd)
}

// -----------------------------------------------------------------------------
// Export Helpers
// -----------------------------------------------------------------------------

var CoinCmd = coinRootCmd

func RegisterCoin(root *cobra.Command) { root.AddCommand(CoinCmd) }
