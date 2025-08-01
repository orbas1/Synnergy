package cli

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	lnLedger string
)

func lnMiddleware(cmd *cobra.Command, _ []string) error {
	_ = godotenv.Load()
	if lp, _ := cmd.Flags().GetString("ledger"); lp != "" {
		lnLedger = lp
	} else if env := os.Getenv("LEDGER_PATH"); env != "" {
		lnLedger = env
	} else {
		lnLedger = "state.db"
	}
	led, err := core.NewLedger(core.LedgerConfig{WALPath: lnLedger})
	if err != nil {
		return err
	}
	core.InitLightning(led)
	return nil
}

func parseAddr(hexStr string) (core.Address, error) {
	b, err := hex.DecodeString(hexStr)
	var a core.Address
	if err != nil || len(b) != len(a) {
		return a, errors.New("address hex invalid")
	}
	copy(a[:], b)
	return a, nil
}

func lnOpen(cmd *cobra.Command, args []string) error {
	if len(args) < 5 {
		return errors.New("usage: lnode open <addrA> <addrB> <token> <amtA> <amtB>")
	}
	a, err := parseAddr(args[0])
	if err != nil {
		return err
	}
	b, err := parseAddr(args[1])
	if err != nil {
		return err
	}
	tok, err := strconv.ParseUint(args[2], 10, 32)
	if err != nil {
		return err
	}
	amtA, err := strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		return err
	}
	amtB, err := strconv.ParseUint(args[4], 10, 64)
	if err != nil {
		return err
	}
	id, err := core.Lightning().OpenChannel(a, b, core.TokenID(tok), amtA, amtB)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%x\n", id)
	return nil
}

func lnPay(cmd *cobra.Command, args []string) error {
	if len(args) < 3 {
		return errors.New("usage: lnode pay <chanID> <from> <amt>")
	}
	idBytes, err := hex.DecodeString(args[0])
	if err != nil || len(idBytes) != 32 {
		return errors.New("invalid id")
	}
	var id core.LightningChannelID
	copy(id[:], idBytes)
	from, err := parseAddr(args[1])
	if err != nil {
		return err
	}
	amt, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return err
	}
	return core.Lightning().RoutePayment(id, from, amt)
}

func lnClose(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("usage: lnode close <chanID>")
	}
	idBytes, err := hex.DecodeString(args[0])
	if err != nil || len(idBytes) != 32 {
		return errors.New("invalid id")
	}
	var id core.LightningChannelID
	copy(id[:], idBytes)
	return core.Lightning().CloseChannel(id)
}

func lnList(cmd *cobra.Command, args []string) error {
	list := core.Lightning().ListChannels()
	for _, ch := range list {
		fmt.Fprintf(cmd.OutOrStdout(), "%x %x %d %d\n", ch.ID, ch.Token, ch.BalanceA, ch.BalanceB)
	}
	return nil
}

var lnCmd = &cobra.Command{
	Use:               "lnode",
	Short:             "Lightning node utilities",
	PersistentPreRunE: lnMiddleware,
}

func init() {
	lnCmd.PersistentFlags().String("ledger", "", "path to ledger db")
	lnCmd.AddCommand(&cobra.Command{Use: "open", RunE: lnOpen}, &cobra.Command{Use: "pay", RunE: lnPay}, &cobra.Command{Use: "close", RunE: lnClose}, &cobra.Command{Use: "list", RunE: lnList})
}

var LightningCmd = lnCmd
