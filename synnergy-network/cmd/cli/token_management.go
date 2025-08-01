package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	tmOnce sync.Once
	tmMgr  *core.TokenManager
)

func tmInit(cmd *cobra.Command, _ []string) error {
	var err error
	tmOnce.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		led, e := core.OpenLedger(path)
		if e != nil {
			err = e
			return
		}
		gas := core.NewFlatGasCalculator()
		tmMgr = core.NewTokenManager(led, gas)
	})
	return err
}

func tmParseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func tmHandleCreate(cmd *cobra.Command, _ []string) error {
	name, _ := cmd.Flags().GetString("name")
	symbol, _ := cmd.Flags().GetString("symbol")
	dec, _ := cmd.Flags().GetUint8("dec")
	std, _ := cmd.Flags().GetUint("standard")
	ownerStr, _ := cmd.Flags().GetString("owner")
	supply, _ := cmd.Flags().GetUint64("supply")

	owner, err := tmParseAddr(ownerStr)
	if err != nil {
		return err
	}

	meta := core.Metadata{Name: name, Symbol: symbol, Decimals: dec, Standard: byte(std)}
	id, err := tmMgr.Create(meta, map[core.Address]uint64{owner: supply})
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "token %s created with ID %d\n", symbol, id)
	return nil
}

func tmHandleBalance(cmd *cobra.Command, args []string) error {
	id64, _ := strconv.ParseUint(args[0], 10, 32)
	addr, err := tmParseAddr(args[1])
	if err != nil {
		return err
	}
	bal, err := tmMgr.BalanceOf(core.TokenID(id64), addr)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%d\n", bal)
	return nil
}

func tmHandleTransfer(cmd *cobra.Command, args []string) error {
	id64, _ := strconv.ParseUint(args[0], 10, 32)
	fromStr, _ := cmd.Flags().GetString("from")
	toStr, _ := cmd.Flags().GetString("to")
	amt, _ := cmd.Flags().GetUint64("amt")

	from, err := tmParseAddr(fromStr)
	if err != nil {
		return err
	}
	to, err := tmParseAddr(toStr)
	if err != nil {
		return err
	}
	if err := tmMgr.Transfer(core.TokenID(id64), from, to, amt); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "transfer ok ✔")
	return nil
}

var tokenMgmtCmd = &cobra.Command{
	Use:               "token_management",
	Short:             "High level token management",
	PersistentPreRunE: tmInit,
}

var tmCreateCmd = &cobra.Command{Use: "create", Short: "Create a token", RunE: tmHandleCreate}
var tmBalCmd = &cobra.Command{Use: "balance <id> <addr>", Short: "Token balance", Args: cobra.ExactArgs(2), RunE: tmHandleBalance}
var tmTransferCmd = &cobra.Command{Use: "transfer <id>", Short: "Transfer tokens", Args: cobra.ExactArgs(1), RunE: tmHandleTransfer}

func init() {
	tmCreateCmd.Flags().String("name", "", "token name")
	tmCreateCmd.Flags().String("symbol", "", "symbol")
	tmCreateCmd.Flags().Uint8("dec", 18, "decimals")
	tmCreateCmd.Flags().Uint("standard", 0x14, "standard byte")
	tmCreateCmd.Flags().String("owner", "", "initial owner")
	tmCreateCmd.Flags().Uint64("supply", 0, "initial supply")
	tmCreateCmd.MarkFlagRequired("name")
	tmCreateCmd.MarkFlagRequired("symbol")
	tmCreateCmd.MarkFlagRequired("owner")

	tmTransferCmd.Flags().String("from", "", "sender")
	tmTransferCmd.Flags().String("to", "", "recipient")
	tmTransferCmd.Flags().Uint64("amt", 0, "amount")
	tmTransferCmd.MarkFlagRequired("from")
	tmTransferCmd.MarkFlagRequired("to")
	tmTransferCmd.MarkFlagRequired("amt")

	tokenMgmtCmd.AddCommand(tmCreateCmd, tmBalCmd, tmTransferCmd)
}

var TokenMgmtCmd = tokenMgmtCmd

func RegisterTokenMgmt(root *cobra.Command) { root.AddCommand(TokenMgmtCmd) }
