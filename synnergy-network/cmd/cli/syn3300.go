package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

func etfResolve(idOrSym string) (*core.SYN3300Token, error) {
	for _, t := range core.ListSYN3300() {
		if fmt.Sprintf("%d", t.ID()) == idOrSym || t.Meta().Symbol == idOrSym {
			return t, nil
		}
	}
	return nil, fmt.Errorf("etf token not found")
}

func etfHandleInfo(cmd *cobra.Command, args []string) error {
	tok, err := etfResolve(args[0])
	if err != nil {
		return err
	}
	info := tok.GetETFInfo()
	fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%d\t%d\t%d\n", info.ETFID, info.Name, info.TotalShares, info.AvailableShares, info.CurrentPrice)
	return nil
}

func etfHandleUpdate(cmd *cobra.Command, args []string) error {
	tok, err := etfResolve(args[0])
	if err != nil {
		return err
	}
	price, _ := cmd.Flags().GetUint64("price")
	tok.UpdatePrice(price)
	fmt.Fprintln(cmd.OutOrStdout(), "updated ✔")
	return nil
}

func etfHandleMint(cmd *cobra.Command, args []string) error {
	tok, err := etfResolve(args[0])
	if err != nil {
		return err
	}
	to, err := tokParseAddr(args[1])
	if err != nil {
		return err
	}
	shares, _ := cmd.Flags().GetUint64("shares")
	return tok.FractionalMint(to, shares)
}

func etfHandleBurn(cmd *cobra.Command, args []string) error {
	tok, err := etfResolve(args[0])
	if err != nil {
		return err
	}
	from, err := tokParseAddr(args[1])
	if err != nil {
		return err
	}
	shares, _ := cmd.Flags().GetUint64("shares")
	return tok.FractionalBurn(from, shares)
}

var syn3300Cmd = &cobra.Command{Use: "syn3300", Short: "Manage SYN3300 ETF tokens"}
var etfInfoCmd = &cobra.Command{Use: "info <id>", Args: cobra.ExactArgs(1), RunE: etfHandleInfo, Short: "ETF info"}
var etfUpdateCmd = &cobra.Command{Use: "update <id>", Args: cobra.ExactArgs(1), RunE: etfHandleUpdate, Short: "Update price"}
var etfMintCmd = &cobra.Command{Use: "mint <id> <to>", Args: cobra.ExactArgs(2), RunE: etfHandleMint, Short: "Mint shares"}
var etfBurnCmd = &cobra.Command{Use: "burn <id> <from>", Args: cobra.ExactArgs(2), RunE: etfHandleBurn, Short: "Burn shares"}

func init() {
	etfUpdateCmd.Flags().Uint64("price", 0, "price")
	etfMintCmd.Flags().Uint64("shares", 0, "shares")
	etfBurnCmd.Flags().Uint64("shares", 0, "shares")
	syn3300Cmd.AddCommand(etfInfoCmd, etfUpdateCmd, etfMintCmd, etfBurnCmd)
}

var SYN3300Cmd = syn3300Cmd
