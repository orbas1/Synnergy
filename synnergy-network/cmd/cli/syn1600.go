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
	mrOnce   sync.Once
	mrLedger *core.Ledger
	mrToken  *core.SYN1600Token
)

func mrInit(cmd *cobra.Command, _ []string) error {
	var err error
	mrOnce.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		mrLedger, err = core.OpenLedger(path)
		if err != nil {
			return
		}
		// assume canonical token exists
		tok, ok := core.GetToken(core.TokenID(deriveID(core.StdSYN1600)))
		if !ok {
			err = fmt.Errorf("SYN1600 token not found")
			return
		}
		var okc bool
		mrToken, okc = tok.(*core.SYN1600Token)
		if !okc {
			// token created via factory without custom struct
			bt := tok.(*core.BaseToken)
			mrToken = &core.SYN1600Token{BaseToken: bt}
		}
	})
	return err
}

func mrHandleInfo(cmd *cobra.Command, _ []string) error {
	b, _ := json.MarshalIndent(mrToken.Info, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(b))
	return nil
}

func mrHandleUpdate(cmd *cobra.Command, _ []string) error {
	title, _ := cmd.Flags().GetString("title")
	artist, _ := cmd.Flags().GetString("artist")
	album, _ := cmd.Flags().GetString("album")
	mrToken.UpdateInfo(core.MusicInfo{SongTitle: title, Artist: artist, Album: album})
	fmt.Fprintln(cmd.OutOrStdout(), "updated ✔")
	return nil
}

func mrHandleDistribute(cmd *cobra.Command, args []string) error {
	amt, _ := cmd.Flags().GetUint64("amt")
	if err := mrToken.DistributeRoyalties(amt); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "distributed ✔")
	return nil
}

var syn1600Cmd = &cobra.Command{
	Use:               "syn1600",
	Short:             "Manage SYN1600 music royalty tokens",
	PersistentPreRunE: mrInit,
}

var mrInfoCmd = &cobra.Command{Use: "info", Short: "Show music info", RunE: mrHandleInfo}
var mrUpdateCmd = &cobra.Command{Use: "update", Short: "Update music info", RunE: mrHandleUpdate}
var mrDistributeCmd = &cobra.Command{Use: "distribute", Short: "Distribute royalties", RunE: mrHandleDistribute}

func init() {
	mrUpdateCmd.Flags().String("title", "", "song title")
	mrUpdateCmd.Flags().String("artist", "", "artist")
	mrUpdateCmd.Flags().String("album", "", "album")
	mrDistributeCmd.Flags().Uint64("amt", 0, "amount")
	mrDistributeCmd.MarkFlagRequired("amt")

	syn1600Cmd.AddCommand(mrInfoCmd, mrUpdateCmd, mrDistributeCmd)
}

// MusicRoyaltyCmd exposes commands for RegisterRoutes
var MusicRoyaltyCmd = syn1600Cmd
