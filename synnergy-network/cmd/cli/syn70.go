package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	Tokens "synnergy-network/core/Tokens"
)

var syn70 *Tokens.SYN70Token

func ensureSYN70(cmd *cobra.Command, _ []string) error {
	if syn70 != nil {
		return nil
	}
	syn70 = Tokens.NewSYN70()
	return nil
}

var syn70Cmd = &cobra.Command{
	Use:               "syn70",
	Short:             "Manage SYN70 in-game assets",
	PersistentPreRunE: ensureSYN70,
}

var syn70RegisterCmd = &cobra.Command{
	Use:   "register <id> <owner> <name> <game>",
	Short: "Register a new asset",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		owner, err := parseSYN70Addr(args[1])
		if err != nil {
			return err
		}
		name := args[2]
		game := args[3]
		asset := &Tokens.SYN70Asset{TokenID: 0, Name: name, Owner: owner, GameID: game}
		return syn70.RegisterAsset(id, asset)
	},
}

var syn70TransferCmd = &cobra.Command{
	Use:   "transfer <id> <newOwner>",
	Short: "Transfer asset ownership",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, err := parseSYN70Addr(args[1])
		if err != nil {
			return err
		}
		return syn70.TransferAsset(args[0], owner)
	},
}

var syn70AttrCmd = &cobra.Command{
	Use:   "setattr <id> <key> <value>",
	Short: "Set asset attribute",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return syn70.UpdateAttributes(args[0], map[string]string{args[1]: args[2]})
	},
}

var syn70AchCmd = &cobra.Command{
	Use:   "achievement <id> <name>",
	Short: "Record achievement",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return syn70.RecordAchievement(args[0], args[1])
	},
}

var syn70InfoCmd = &cobra.Command{
	Use:   "info <id>",
	Short: "Show asset info",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, ok := syn70.GetAsset(args[0])
		if !ok {
			return fmt.Errorf("asset not found")
		}
		b, _ := json.MarshalIndent(a, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var syn70ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List assets",
	RunE: func(cmd *cobra.Command, args []string) error {
		list := syn70.ListAssets()
		b, _ := json.MarshalIndent(list, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

func init() {
	syn70Cmd.AddCommand(syn70RegisterCmd, syn70TransferCmd, syn70AttrCmd, syn70AchCmd, syn70InfoCmd, syn70ListCmd)
}

var SYN70Cmd = syn70Cmd

func RegisterSYN70(root *cobra.Command) { root.AddCommand(SYN70Cmd) }

func parseSYN70Addr(h string) (Tokens.Address, error) {
	var a Tokens.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}
