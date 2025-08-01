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
	gameOnce sync.Once
)

func gameInit(cmd *cobra.Command, _ []string) error {
	var err error
	gameOnce.Do(func() {
		_ = godotenv.Load()
		path, _ := cmd.Flags().GetString("ledger")
		if path == "" {
			path = os.Getenv("LEDGER_PATH")
		}
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		led, e := core.OpenLedger(path)
		if e != nil {
			err = e
			return
		}
		core.InitGaming(led)
	})
	return err
}

func gameCreate(cmd *cobra.Command, _ []string) error {
	creatorStr, _ := cmd.Flags().GetString("creator")
	stake, _ := cmd.Flags().GetUint64("stake")
	addr, err := core.ParseAddress(creatorStr)
	if err != nil {
		return err
	}
	g, err := core.CreateGame(addr, stake)
	if err != nil {
		return err
	}
	enc, _ := json.MarshalIndent(g, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(enc))
	return nil
}

func gameJoin(cmd *cobra.Command, args []string) error {
	playerStr, _ := cmd.Flags().GetString("player")
	addr, err := core.ParseAddress(playerStr)
	if err != nil {
		return err
	}
	if err := core.JoinGame(args[0], addr); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "joined ✔")
	return nil
}

func gameFinish(cmd *cobra.Command, args []string) error {
	winnerStr, _ := cmd.Flags().GetString("winner")
	addr, err := core.ParseAddress(winnerStr)
	if err != nil {
		return err
	}
	g, err := core.FinishGame(args[0], addr)
	if err != nil {
		return err
	}
	enc, _ := json.MarshalIndent(g, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(enc))
	return nil
}

func gameGet(cmd *cobra.Command, args []string) error {
	g, err := core.GetGame(args[0])
	if err != nil {
		return err
	}
	enc, _ := json.MarshalIndent(g, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(enc))
	return nil
}

func gameList(cmd *cobra.Command, _ []string) error {
	games, err := core.ListGames()
	if err != nil {
		return err
	}
	enc, _ := json.MarshalIndent(games, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(enc))
	return nil
}

var gamingCmd = &cobra.Command{
	Use:               "gaming",
	Short:             "Manage simple on-chain games",
	PersistentPreRunE: gameInit,
}

var gameCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new game",
	RunE:  gameCreate,
}

var gameJoinCmd = &cobra.Command{
	Use:   "join <id>",
	Short: "Join an existing game",
	Args:  cobra.ExactArgs(1),
	RunE:  gameJoin,
}

var gameFinishCmd = &cobra.Command{
	Use:   "finish <id>",
	Short: "Finish a game and pay out",
	Args:  cobra.ExactArgs(1),
	RunE:  gameFinish,
}

var gameGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show game info",
	Args:  cobra.ExactArgs(1),
	RunE:  gameGet,
}

var gameListCmd = &cobra.Command{
	Use:   "list",
	Short: "List games",
	Args:  cobra.NoArgs,
	RunE:  gameList,
}

func init() {
	gamingCmd.PersistentFlags().String("ledger", "", "path to ledger")

	gameCreateCmd.Flags().String("creator", "", "creator address")
	gameCreateCmd.MarkFlagRequired("creator")
	gameCreateCmd.Flags().Uint64("stake", 0, "stake amount")

	gameJoinCmd.Flags().String("player", "", "player address")
	gameJoinCmd.MarkFlagRequired("player")

	gameFinishCmd.Flags().String("winner", "", "winner address")
	gameFinishCmd.MarkFlagRequired("winner")

	gamingCmd.AddCommand(gameCreateCmd, gameJoinCmd, gameFinishCmd, gameGetCmd, gameListCmd)
}

// GamingCmd exported for index.go
var GamingCmd = gamingCmd
