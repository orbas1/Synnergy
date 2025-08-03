package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	syn721Once sync.Once
	syn721Mgr  *core.TokenManager
)

func syn721Init(cmd *cobra.Command, _ []string) error {
	var err error
	syn721Once.Do(func() {
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
		gas := core.NewFlatGasCalculator(core.DefaultGasPrice)
		syn721Mgr = core.NewTokenManager(led, gas)
	})
	return err
}

func parseAddr721(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func syn721HandleMint(cmd *cobra.Command, args []string) error {
	id64, _ := cmd.Flags().GetUint("id")
	toStr, _ := cmd.Flags().GetString("to")
	uri, _ := cmd.Flags().GetString("uri")
	meta, _ := cmd.Flags().GetString("data")
	to, err := parseAddr721(toStr)
	if err != nil {
		return err
	}
	_, err = syn721Mgr.Mint721(core.TokenID(id64), to, core.SYN721Metadata{URI: uri, Data: meta})
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "minted ✔")
	return nil
}

func syn721HandleTransfer(cmd *cobra.Command, args []string) error {
	id64, _ := cmd.Flags().GetUint("id")
	tokenID, _ := cmd.Flags().GetUint64("token")
	fromStr, _ := cmd.Flags().GetString("from")
	toStr, _ := cmd.Flags().GetString("to")
	from, err := parseAddr721(fromStr)
	if err != nil {
		return err
	}
	to, err := parseAddr721(toStr)
	if err != nil {
		return err
	}
	return syn721Mgr.Transfer721(core.TokenID(id64), from, to, tokenID)
}

func syn721HandleMetadata(cmd *cobra.Command, args []string) error {
	id64, _ := cmd.Flags().GetUint("id")
	tokenID, _ := cmd.Flags().GetUint64("token")
	m, err := syn721Mgr.Metadata721(core.TokenID(id64), tokenID)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", m.Data)
	return nil
}

var syn721Cmd = &cobra.Command{
	Use:               "syn721",
	Short:             "Manage SYN721 NFTs",
	PersistentPreRunE: syn721Init,
}

var syn721MintCmd = &cobra.Command{Use: "mint", RunE: syn721HandleMint}
var syn721TransferCmd = &cobra.Command{Use: "transfer", RunE: syn721HandleTransfer}
var syn721MetaCmd = &cobra.Command{Use: "metadata", RunE: syn721HandleMetadata}

func init() {
	syn721MintCmd.Flags().Uint("id", 0, "token id")
	syn721MintCmd.Flags().String("to", "", "recipient")
	syn721MintCmd.Flags().String("uri", "", "uri")
	syn721MintCmd.Flags().String("data", "", "data")
	syn721MintCmd.MarkFlagRequired("id")
	syn721MintCmd.MarkFlagRequired("to")

	syn721TransferCmd.Flags().Uint("id", 0, "token id")
	syn721TransferCmd.Flags().Uint64("token", 0, "nft id")
	syn721TransferCmd.Flags().String("from", "", "from")
	syn721TransferCmd.Flags().String("to", "", "to")
	syn721TransferCmd.MarkFlagRequired("id")
	syn721TransferCmd.MarkFlagRequired("token")
	syn721TransferCmd.MarkFlagRequired("from")
	syn721TransferCmd.MarkFlagRequired("to")

	syn721MetaCmd.Flags().Uint("id", 0, "token id")
	syn721MetaCmd.Flags().Uint64("token", 0, "nft id")
	syn721MetaCmd.MarkFlagRequired("id")
	syn721MetaCmd.MarkFlagRequired("token")

	syn721Cmd.AddCommand(syn721MintCmd, syn721TransferCmd, syn721MetaCmd)
}

var SYN721Cmd = syn721Cmd

func RegisterSYN721(root *cobra.Command) { root.AddCommand(SYN721Cmd) }
