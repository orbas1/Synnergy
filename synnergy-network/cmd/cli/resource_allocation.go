package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	resOnce   sync.Once
	resMgr    *core.ResourceManager
	resLogger = logrus.StandardLogger()
)

func resInit(cmd *cobra.Command, _ []string) error {
	var err error
	resOnce.Do(func() {
		_ = godotenv.Load()
		lvl := os.Getenv("LOG_LEVEL")
		if lvl == "" {
			lvl = "info"
		}
		lv, e := logrus.ParseLevel(lvl)
		if e != nil {
			err = e
			return
		}
		resLogger.SetLevel(lv)
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
		resMgr = core.NewResourceManager(led)
	})
	return err
}

func resParseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func resHandleSet(cmd *cobra.Command, args []string) error {
	addr, err := resParseAddr(args[0])
	if err != nil {
		return err
	}
	limit, _ := cmd.Flags().GetUint64("limit")
	if err := resMgr.SetLimit(addr, limit); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "ok")
	return nil
}

func resHandleGet(cmd *cobra.Command, args []string) error {
	addr, err := resParseAddr(args[0])
	if err != nil {
		return err
	}
	lim, err := resMgr.GetLimit(addr)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%d\n", lim)
	return nil
}

func resHandleList(cmd *cobra.Command, _ []string) error {
	m, err := resMgr.ListLimits()
	if err != nil {
		return err
	}
	for a, l := range m {
		fmt.Fprintf(cmd.OutOrStdout(), "%x\t%d\n", a, l)
	}
	return nil
}

func resHandleConsume(cmd *cobra.Command, args []string) error {
	addr, err := resParseAddr(args[0])
	if err != nil {
		return err
	}
	amt, _ := cmd.Flags().GetUint64("amt")
	if err := resMgr.Consume(addr, amt); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "ok")
	return nil
}

func resHandleTransfer(cmd *cobra.Command, args []string) error {
	from, err := resParseAddr(args[0])
	if err != nil {
		return err
	}
	to, err := resParseAddr(args[1])
	if err != nil {
		return err
	}
	amt, _ := cmd.Flags().GetUint64("amt")
	if err := resMgr.TransferLimit(from, to, amt); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "ok")
	return nil
}

var resCmd = &cobra.Command{
	Use:               "resource_allocation",
	Short:             "Manage gas allocation per contract address",
	PersistentPreRunE: resInit,
}

var resSetCmd = &cobra.Command{Use: "set <addr>", Short: "Set gas limit", Args: cobra.ExactArgs(1), RunE: resHandleSet}
var resGetCmd = &cobra.Command{Use: "get <addr>", Short: "Get gas limit", Args: cobra.ExactArgs(1), RunE: resHandleGet}
var resListCmd = &cobra.Command{Use: "list", Short: "List limits", Args: cobra.NoArgs, RunE: resHandleList}
var resConsumeCmd = &cobra.Command{Use: "consume <addr>", Short: "Consume gas", Args: cobra.ExactArgs(1), RunE: resHandleConsume}
var resTransferCmd = &cobra.Command{Use: "transfer <from> <to>", Short: "Transfer limit", Args: cobra.ExactArgs(2), RunE: resHandleTransfer}

func init() {
	resSetCmd.Flags().Uint64("limit", 0, "gas limit")
	resSetCmd.MarkFlagRequired("limit")
	resConsumeCmd.Flags().Uint64("amt", 0, "amount")
	resConsumeCmd.MarkFlagRequired("amt")
	resTransferCmd.Flags().Uint64("amt", 0, "amount")
	resTransferCmd.MarkFlagRequired("amt")

	resCmd.AddCommand(resSetCmd, resGetCmd, resListCmd, resConsumeCmd, resTransferCmd)
}

var ResourceCmd = resCmd
