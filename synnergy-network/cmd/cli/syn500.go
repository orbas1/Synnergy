package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"time"

	core "synnergy-network/core"
)

var (
	syn500Once   sync.Once
	syn500Tok    *core.SYN500Token
	syn500Ledger *core.Ledger
)

func syn500Init(cmd *cobra.Command, _ []string) error {
	var err error
	syn500Once.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		syn500Ledger, err = core.OpenLedger(path)
		if err != nil {
			return
		}
	})
	return err
}

func syn500ParseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func syn500HandleCreate(cmd *cobra.Command, _ []string) error {
	name, _ := cmd.Flags().GetString("name")
	symbol, _ := cmd.Flags().GetString("symbol")
	dec, _ := cmd.Flags().GetUint8("dec")
	ownerStr, _ := cmd.Flags().GetString("owner")
	supply, _ := cmd.Flags().GetUint64("supply")
	owner, err := syn500ParseAddr(ownerStr)
	if err != nil {
		return err
	}
	meta := core.Metadata{Name: name, Symbol: symbol, Decimals: dec, Standard: core.StdSYN500}
	mgr := core.NewTokenManager(syn500Ledger, core.NewFlatGasCalculator())
	tok, err := mgr.CreateSYN500(meta, map[core.Address]uint64{owner: supply})
	if err != nil {
		return err
	}
	syn500Tok = tok
	fmt.Fprintf(cmd.OutOrStdout(), "SYN500 token created with ID %d\n", tok.ID())
	return nil
}

func syn500HandleGrant(cmd *cobra.Command, args []string) error {
	if syn500Tok == nil {
		return fmt.Errorf("token not loaded")
	}
	addr, err := syn500ParseAddr(args[0])
	if err != nil {
		return err
	}
	tier, _ := cmd.Flags().GetUint8("tier")
	max, _ := cmd.Flags().GetUint64("max")
	syn500Tok.GrantAccess(addr, core.ServiceTier(tier), max, time.Time{})
	fmt.Fprintln(cmd.OutOrStdout(), "access granted")
	return nil
}

func syn500HandleUsage(cmd *cobra.Command, args []string) error {
	if syn500Tok == nil {
		return fmt.Errorf("token not loaded")
	}
	addr, err := syn500ParseAddr(args[0])
	if err != nil {
		return err
	}
	if err := syn500Tok.RecordUsage(addr, 1); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "usage recorded")
	return nil
}

var syn500Cmd = &cobra.Command{
	Use:               "syn500",
	Short:             "Manage SYN500 utility tokens",
	PersistentPreRunE: syn500Init,
}

var syn500CreateCmd = &cobra.Command{Use: "create", Short: "Create token", RunE: syn500HandleCreate}
var syn500GrantCmd = &cobra.Command{Use: "grant <addr>", Short: "Grant access", Args: cobra.ExactArgs(1), RunE: syn500HandleGrant}
var syn500UseCmd = &cobra.Command{Use: "use <addr>", Short: "Record usage", Args: cobra.ExactArgs(1), RunE: syn500HandleUsage}

func init() {
	syn500CreateCmd.Flags().String("name", "", "name")
	syn500CreateCmd.Flags().String("symbol", "", "symbol")
	syn500CreateCmd.Flags().Uint8("dec", 18, "decimals")
	syn500CreateCmd.Flags().String("owner", "", "owner")
	syn500CreateCmd.Flags().Uint64("supply", 0, "supply")
	syn500CreateCmd.MarkFlagRequired("name")
	syn500CreateCmd.MarkFlagRequired("symbol")
	syn500CreateCmd.MarkFlagRequired("owner")

	syn500GrantCmd.Flags().Uint8("tier", 0, "tier")
	syn500GrantCmd.Flags().Uint64("max", 0, "max usage")

	syn500Cmd.AddCommand(syn500CreateCmd, syn500GrantCmd, syn500UseCmd)
}

var Syn500Cmd = syn500Cmd

func RegisterSyn500(root *cobra.Command) { root.AddCommand(Syn500Cmd) }
