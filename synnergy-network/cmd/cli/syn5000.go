package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	syn5Once sync.Once
	syn5Tok  *core.SYN5000Token
)

func syn5Init(cmd *cobra.Command, _ []string) error {
	var err error
	syn5Once.Do(func() {
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
		core.InitSYN5000(led)
		for _, t := range core.GetRegistryTokens() {
			if t.Meta().Standard == core.StdSYN5000 {
				syn5Tok = core.NewSYN5000(t)
				break
			}
		}
		if syn5Tok == nil {
			meta := core.Metadata{Name: "Synnergy Gambling", Symbol: "SYN-GMBL", Decimals: 0, Standard: core.StdSYN5000}
			id, e := core.NewTokenManager(led, core.NewFlatGasCalculator()).Create(meta, map[core.Address]uint64{})
			if e != nil {
				err = e
				return
			}
			if tok, ok := core.GetToken(id); ok {
				syn5Tok = core.NewSYN5000(tok.(*core.BaseToken))
			} else {
				err = fmt.Errorf("token create failed")
			}
		}
	})
	return err
}

func syn5Place(cmd *cobra.Command, _ []string) error {
	playerStr, _ := cmd.Flags().GetString("player")
	gameType, _ := cmd.Flags().GetString("type")
	amt, _ := cmd.Flags().GetUint64("amt")
	addr, err := core.ParseAddress(playerStr)
	if err != nil {
		return err
	}
	b, err := syn5Tok.PlaceBet(addr, gameType, amt)
	if err != nil {
		return err
	}
	enc, _ := json.MarshalIndent(b, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(enc))
	return nil
}

func syn5Resolve(cmd *cobra.Command, args []string) error {
	winnerStr, _ := cmd.Flags().GetString("winner")
	outcome, _ := cmd.Flags().GetString("outcome")
	addr, err := core.ParseAddress(winnerStr)
	if err != nil {
		return err
	}
	if err := syn5Tok.ResolveBet(args[0], outcome, addr); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "resolved ✔")
	return nil
}

func syn5Get(cmd *cobra.Command, args []string) error {
	b, err := syn5Tok.GetBet(args[0])
	if err != nil {
		return err
	}
	enc, _ := json.MarshalIndent(b, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(enc))
	return nil
}

func syn5List(cmd *cobra.Command, _ []string) error {
	list, err := syn5Tok.ListBets()
	if err != nil {
		return err
	}
	enc, _ := json.MarshalIndent(list, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(enc))
	return nil
}

var syn5Cmd = &cobra.Command{
	Use:               "syn5000",
	Short:             "Interact with SYN5000 gambling tokens",
	PersistentPreRunE: syn5Init,
}

var syn5PlaceCmd = &cobra.Command{Use: "place", Short: "Place a bet", RunE: syn5Place}
var syn5ResolveCmd = &cobra.Command{Use: "resolve <id>", Short: "Resolve a bet", Args: cobra.ExactArgs(1), RunE: syn5Resolve}
var syn5GetCmd = &cobra.Command{Use: "get <id>", Short: "Get bet", Args: cobra.ExactArgs(1), RunE: syn5Get}
var syn5ListCmd = &cobra.Command{Use: "list", Short: "List bets", RunE: syn5List}

func init() {
	syn5PlaceCmd.Flags().String("player", "", "player address")
	syn5PlaceCmd.Flags().String("type", "game", "game type")
	syn5PlaceCmd.Flags().Uint64("amt", 0, "amount")
	syn5PlaceCmd.MarkFlagRequired("player")
	syn5PlaceCmd.MarkFlagRequired("amt")

	syn5ResolveCmd.Flags().String("winner", "", "winner address")
	syn5ResolveCmd.Flags().String("outcome", "", "outcome")
	syn5ResolveCmd.MarkFlagRequired("winner")
	syn5ResolveCmd.MarkFlagRequired("outcome")

	syn5Cmd.AddCommand(syn5PlaceCmd, syn5ResolveCmd, syn5GetCmd, syn5ListCmd)
}

// Syn5000Cmd exported for index.go
var Syn5000Cmd = syn5Cmd
