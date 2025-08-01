package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"synnergy-network/core"
)

var (
	agriOnce bool
	agriTok  *core.Syn4900Token
)

func agriInit(cmd *cobra.Command, _ []string) error {
	if agriOnce {
		return nil
	}
	_ = godotenv.Load()
	for _, t := range core.GetRegistryTokens() {
		if m := t.Meta(); m.Standard == core.StdSYN4900 {
			if tok, ok := core.GetSyn4900(t.ID()); ok {
				agriTok = tok
				agriOnce = true
				return nil
			}
		}
	}
	return fmt.Errorf("SYN4900 token not found")
}

func parseAddrString(addr string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(addr, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func agriRegister(cmd *cobra.Command, args []string) error {
	var asset core.AgriculturalAsset
	asset.ID = args[0]
	asset.AssetType = args[1]
	qty, _ := cmd.Flags().GetUint64("qty")
	asset.Quantity = qty
	asset.Origin, _ = cmd.Flags().GetString("origin")
	asset.Status = "created"
	agriTok.RegisterAsset(asset)
	fmt.Fprintln(cmd.OutOrStdout(), "asset registered")
	return nil
}

func agriTransfer(cmd *cobra.Command, args []string) error {
	id := args[0]
	fromStr, _ := cmd.Flags().GetString("from")
	toStr, _ := cmd.Flags().GetString("to")
	qty, _ := cmd.Flags().GetUint64("qty")
	from, err := parseAddrString(fromStr)
	if err != nil {
		return err
	}
	to, err := parseAddrString(toStr)
	if err != nil {
		return err
	}
	return agriTok.TransferAsset(id, from, to, qty)
}

func agriInvest(cmd *cobra.Command, args []string) error {
	investorStr, _ := cmd.Flags().GetString("addr")
	amt, _ := cmd.Flags().GetUint64("amt")
	inv, err := parseAddrString(investorStr)
	if err != nil {
		return err
	}
	agriTok.RecordInvestment(inv, amt)
	fmt.Fprintln(cmd.OutOrStdout(), "investment recorded")
	return nil
}

func agriGetInv(cmd *cobra.Command, args []string) error {
	investorStr := args[0]
	inv, err := parseAddrString(investorStr)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%d\n", agriTok.InvestmentOf(inv))
	return nil
}

var agricultureCmd = &cobra.Command{
	Use:               "agriculture",
	Short:             "Manage SYN4900 agricultural tokens",
	PersistentPreRunE: agriInit,
}

var agriRegisterCmd = &cobra.Command{
	Use:   "register <id> <type>",
	Short: "Register asset",
	Args:  cobra.ExactArgs(2),
	RunE:  agriRegister,
}

var agriTransferCmd = &cobra.Command{
	Use:   "transfer <assetID>",
	Short: "Transfer asset",
	Args:  cobra.ExactArgs(1),
	RunE:  agriTransfer,
}

var agriInvestCmd = &cobra.Command{
	Use:   "invest",
	Short: "Record investment",
	Args:  cobra.NoArgs,
	RunE:  agriInvest,
}

var agriGetInvCmd = &cobra.Command{
	Use:   "invest_of <addr>",
	Short: "Get investment amount",
	Args:  cobra.ExactArgs(1),
	RunE:  agriGetInv,
}

func init() {
	agriRegisterCmd.Flags().Uint64("qty", 0, "quantity")
	agriRegisterCmd.Flags().String("origin", "", "origin")

	agriTransferCmd.Flags().String("from", "", "sender")
	agriTransferCmd.Flags().String("to", "", "recipient")
	agriTransferCmd.Flags().Uint64("qty", 0, "quantity")
	agriTransferCmd.MarkFlagRequired("from")
	agriTransferCmd.MarkFlagRequired("to")

	agriInvestCmd.Flags().String("addr", "", "investor address")
	agriInvestCmd.Flags().Uint64("amt", 0, "amount")
	agriInvestCmd.MarkFlagRequired("addr")
	agriInvestCmd.MarkFlagRequired("amt")

	agricultureCmd.AddCommand(agriRegisterCmd, agriTransferCmd, agriInvestCmd, agriGetInvCmd)
}

// AgricultureCmd entry point
var AgricultureCmd = agricultureCmd
