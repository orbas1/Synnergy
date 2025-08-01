package cli

// cmd/cli/user_feedback_system.go - CLI for interacting with the on-chain
// feedback engine. Commands are consolidated under "feedback" and operate
// directly on the ledger using helper functions from the core package.

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	fbOnce   sync.Once
	fbErr    error
	fbEngine *core.FeedbackEngine
)

func fbInit(cmd *cobra.Command, _ []string) error {
	fbOnce.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			fbErr = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		if err := core.InitLedger(path); err != nil {
			fbErr = err
			return
		}
		core.InitFeedback(core.CurrentLedger())
		fbEngine = core.Feedback()
	})
	return fbErr
}

// submit ----------------------------------------------------------------------
var fbSubmitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit user feedback",
	RunE: func(cmd *cobra.Command, _ []string) error {
		userStr, _ := cmd.Flags().GetString("user")
		rating, _ := cmd.Flags().GetUint8("rating")
		msg, _ := cmd.Flags().GetString("message")
		b, err := hex.DecodeString(userStr)
		if err != nil || len(b) != 20 {
			return fmt.Errorf("invalid user address")
		}
		var addr core.Address
		copy(addr[:], b)
		id, err := fbEngine.Submit(addr, rating, msg)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), id)
		return nil
	},
	Args: cobra.NoArgs,
}

// get -------------------------------------------------------------------------
var fbGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get feedback by id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		e, err := fbEngine.Get(args[0])
		if err != nil {
			return err
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(e)
	},
}

// list ------------------------------------------------------------------------
var fbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all feedback entries",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		list, err := fbEngine.List()
		if err != nil {
			return err
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(list)
	},
}

// reward ----------------------------------------------------------------------
var fbRewardCmd = &cobra.Command{
	Use:   "reward <id> --amt=100",
	Short: "Reward a feedback entry with SYNN",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		amt, _ := cmd.Flags().GetUint64("amt")
		return fbEngine.Reward(args[0], amt)
	},
}

// root ------------------------------------------------------------------------
var FeedbackCmd = &cobra.Command{
	Use:               "feedback",
	Short:             "Interact with the feedback system",
	PersistentPreRunE: fbInit,
}

func init() {
	fbSubmitCmd.Flags().String("user", "", "user address")
	fbSubmitCmd.Flags().Uint8("rating", 0, "rating 1-5")
	fbSubmitCmd.Flags().String("message", "", "feedback message")
	fbSubmitCmd.MarkFlagRequired("user")
	fbSubmitCmd.MarkFlagRequired("rating")
	fbSubmitCmd.MarkFlagRequired("message")

	fbRewardCmd.Flags().Uint64("amt", 0, "amount of SYNN to mint")
	fbRewardCmd.MarkFlagRequired("amt")

	FeedbackCmd.AddCommand(fbSubmitCmd, fbGetCmd, fbListCmd, fbRewardCmd)
}
