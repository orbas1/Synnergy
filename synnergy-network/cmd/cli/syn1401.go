package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	synOnce sync.Once
	synMgr  *core.InvestmentManager
)

func synInit(cmd *cobra.Command, _ []string) error {
	var err error
	synOnce.Do(func() {
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
		synMgr = core.NewInvestmentManager(led)
	})
	return err
}

func parseSynAddr(s string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func synHandleIssue(cmd *cobra.Command, _ []string) error {
	id, _ := cmd.Flags().GetString("id")
	ownerStr, _ := cmd.Flags().GetString("owner")
	principal, _ := cmd.Flags().GetUint64("principal")
	rate, _ := cmd.Flags().GetFloat64("rate")
	matDays, _ := cmd.Flags().GetInt("maturity")
	owner, err := parseSynAddr(ownerStr)
	if err != nil {
		return err
	}
	mat := time.Now().Add(time.Duration(matDays) * 24 * time.Hour)
	return synMgr.Issue(id, owner, principal, rate, mat)
}

func synHandleAccrue(cmd *cobra.Command, args []string) error {
	return synMgr.Accrue(args[0])
}

func synHandleRedeem(cmd *cobra.Command, args []string) error {
	toStr, _ := cmd.Flags().GetString("to")
	to, err := parseSynAddr(toStr)
	if err != nil {
		return err
	}
	_, err = synMgr.Redeem(args[0], to)
	return err
}

func synHandleInfo(cmd *cobra.Command, args []string) error {
	rec, ok, err := synMgr.Get(args[0])
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(cmd.OutOrStdout(), "not found")
		return nil
	}
	b, _ := json.MarshalIndent(rec, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(b))
	return nil
}

var syn1401Cmd = &cobra.Command{
	Use:               "syn1401",
	Short:             "Manage SYN1401 investment tokens",
	PersistentPreRunE: synInit,
}

var synIssueCmd = &cobra.Command{Use: "issue", RunE: synHandleIssue}
var synAccrueCmd = &cobra.Command{Use: "accrue <id>", Args: cobra.ExactArgs(1), RunE: synHandleAccrue}
var synRedeemCmd = &cobra.Command{Use: "redeem <id>", Args: cobra.ExactArgs(1), RunE: synHandleRedeem}
var synInfoCmd = &cobra.Command{Use: "info <id>", Args: cobra.ExactArgs(1), RunE: synHandleInfo}

func init() {
	synIssueCmd.Flags().String("id", "", "investment id")
	synIssueCmd.Flags().String("owner", "", "owner address")
	synIssueCmd.Flags().Uint64("principal", 0, "principal amount")
	synIssueCmd.Flags().Float64("rate", 0, "annual interest rate")
	synIssueCmd.Flags().Int("maturity", 0, "maturity in days")
	synIssueCmd.MarkFlagRequired("id")
	synIssueCmd.MarkFlagRequired("owner")
	synIssueCmd.MarkFlagRequired("principal")
	synIssueCmd.MarkFlagRequired("rate")
	synIssueCmd.MarkFlagRequired("maturity")

	synRedeemCmd.Flags().String("to", "", "recipient address")
	synRedeemCmd.MarkFlagRequired("to")

	syn1401Cmd.AddCommand(synIssueCmd, synAccrueCmd, synRedeemCmd, synInfoCmd)
}

var Syn1401Cmd = syn1401Cmd

func RegisterSYN1401(root *cobra.Command) { root.AddCommand(Syn1401Cmd) }
