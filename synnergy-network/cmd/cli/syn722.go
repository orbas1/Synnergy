package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

func syn722Resolve(idStr string) (*core.SYN722Token, error) {
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return nil, err
	}
	tok, ok := core.GetToken(core.TokenID(id64))
	if !ok {
		return nil, fmt.Errorf("token %d not found", id64)
	}
	s, ok := tok.(*core.SYN722Token)
	if !ok {
		return nil, fmt.Errorf("token %d is not SYN722", id64)
	}
	return s, nil
}

func syn722HandleFungible(cmd *cobra.Command, args []string) error {
	t, err := syn722Resolve(args[0])
	if err != nil {
		return err
	}
	t.SetFungible()
	fmt.Fprintln(cmd.OutOrStdout(), "mode set to fungible")
	return nil
}

func syn722HandleNonFungible(cmd *cobra.Command, args []string) error {
	t, err := syn722Resolve(args[0])
	if err != nil {
		return err
	}
	t.SetNonFungible()
	fmt.Fprintln(cmd.OutOrStdout(), "mode set to non-fungible")
	return nil
}

func syn722HandleMode(cmd *cobra.Command, args []string) error {
	t, err := syn722Resolve(args[0])
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%v\n", t.Mode())
	return nil
}

var syn722Cmd = &cobra.Command{Use: "syn722", Short: "Manage SYN722 tokens"}
var syn722FungibleCmd = &cobra.Command{Use: "set-fungible <id>", Args: cobra.ExactArgs(1), RunE: syn722HandleFungible}
var syn722NonFungibleCmd = &cobra.Command{Use: "set-nonfungible <id>", Args: cobra.ExactArgs(1), RunE: syn722HandleNonFungible}
var syn722ModeCmd = &cobra.Command{Use: "mode <id>", Args: cobra.ExactArgs(1), RunE: syn722HandleMode}

func init() {
	syn722Cmd.AddCommand(syn722FungibleCmd, syn722NonFungibleCmd, syn722ModeCmd)
}

var SYN722Cmd = syn722Cmd

func RegisterSYN722(root *cobra.Command) { root.AddCommand(SYN722Cmd) }
