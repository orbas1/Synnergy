package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

func syn1200Token(arg string) (*core.SYN1200Token, error) {
	tok, err := tokResolveToken(arg)
	if err != nil {
		return nil, err
	}
	st, ok := interface{}(tok).(*core.SYN1200Token)
	if !ok {
		return nil, fmt.Errorf("token is not SYN1200 standard")
	}
	return st, nil
}

func syn1200HandleAddBridge(cmd *cobra.Command, args []string) error {
	tok, err := syn1200Token(args[0])
	if err != nil {
		return err
	}
	chain := args[1]
	addrBytes, err := hex.DecodeString(args[2])
	if err != nil || len(addrBytes) != len(core.Address{}) {
		return fmt.Errorf("bad address")
	}
	var addr core.Address
	copy(addr[:], addrBytes)
	tok.AddBridge(chain, addr)
	fmt.Fprintln(cmd.OutOrStdout(), "bridge added")
	return nil
}

func syn1200HandleAtomicSwap(cmd *cobra.Command, args []string) error {
	tok, err := syn1200Token(args[0])
	if err != nil {
		return err
	}
	id, _ := cmd.Flags().GetString("id")
	chain, _ := cmd.Flags().GetString("chain")
	fromStr, _ := cmd.Flags().GetString("from")
	toStr, _ := cmd.Flags().GetString("to")
	amt, _ := cmd.Flags().GetUint64("amt")
	from, err := tokParseAddr(fromStr)
	if err != nil {
		return err
	}
	to, err := tokParseAddr(toStr)
	if err != nil {
		return err
	}
	if err := tok.AtomicSwap(id, chain, from, to, amt); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "swap initiated")
	return nil
}

func syn1200HandleSwapStatus(cmd *cobra.Command, args []string) error {
	tok, err := syn1200Token(args[0])
	if err != nil {
		return err
	}
	rec, ok := tok.GetSwap(args[1])
	if !ok {
		return fmt.Errorf("swap not found")
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%+v\n", *rec)
	return nil
}

var syn1200Cmd = &cobra.Command{Use: "syn1200", Short: "Manage SYN1200 interoperability tokens", PersistentPreRunE: tokInitMiddleware}
var syn1200AddBridgeCmd = &cobra.Command{Use: "add-bridge <token> <chain> <addr>", Short: "Add bridge", Args: cobra.ExactArgs(3), RunE: syn1200HandleAddBridge}
var syn1200SwapCmd = &cobra.Command{Use: "swap <token>", Short: "Atomic swap", Args: cobra.ExactArgs(1), RunE: syn1200HandleAtomicSwap}
var syn1200StatusCmd = &cobra.Command{Use: "status <token> <id>", Short: "Swap status", Args: cobra.ExactArgs(2), RunE: syn1200HandleSwapStatus}

func init() {
	syn1200SwapCmd.Flags().String("id", "", "swap id")
	syn1200SwapCmd.Flags().String("chain", "", "partner chain")
	syn1200SwapCmd.Flags().String("from", "", "from address")
	syn1200SwapCmd.Flags().String("to", "", "to address")
	syn1200SwapCmd.Flags().Uint64("amt", 0, "amount")
	syn1200SwapCmd.MarkFlagRequired("id")
	syn1200SwapCmd.MarkFlagRequired("chain")
	syn1200SwapCmd.MarkFlagRequired("from")
	syn1200SwapCmd.MarkFlagRequired("to")
	syn1200SwapCmd.MarkFlagRequired("amt")

	syn1200Cmd.AddCommand(syn1200AddBridgeCmd, syn1200SwapCmd, syn1200StatusCmd)
}

var Syn1200Cmd = syn1200Cmd

func RegisterSYN1200(root *cobra.Command) { root.AddCommand(Syn1200Cmd) }
