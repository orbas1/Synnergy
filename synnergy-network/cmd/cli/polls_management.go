package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

// PollController wraps core poll management functions for the CLI.
type PollController struct{}

func (PollController) Create(question string, opts []string, creator string, dur time.Duration) (core.Poll, error) {
	addr, err := core.ParseAddress(creator)
	if err != nil {
		return core.Poll{}, err
	}
	return core.CreatePoll(question, opts, addr, dur)
}

func (PollController) Vote(id, voter string, option int) error {
	addr, err := core.ParseAddress(voter)
	if err != nil {
		return err
	}
	return core.VotePoll(id, addr, option)
}

func (PollController) Close(id string) error            { return core.ClosePoll(id) }
func (PollController) Get(id string) (core.Poll, error) { return core.GetPoll(id) }
func (PollController) List() ([]core.Poll, error)       { return core.ListPolls() }

// ----------------------------------------------------------------------
// CLI commands
// ----------------------------------------------------------------------

var pollsCmd = &cobra.Command{Use: "polls", Short: "Manage on-chain polls"}

var pollsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new poll",
	RunE: func(cmd *cobra.Command, args []string) error {
		q, _ := cmd.Flags().GetString("question")
		optStr, _ := cmd.Flags().GetStringSlice("option")
		creator, _ := cmd.Flags().GetString("creator")
		durStr, _ := cmd.Flags().GetString("duration")
		if q == "" || len(optStr) < 2 {
			return fmt.Errorf("question and at least two --option required")
		}
		d, err := time.ParseDuration(durStr)
		if durStr == "" || err != nil {
			d = 72 * time.Hour
		}
		ctrl := PollController{}
		p, err := ctrl.Create(q, optStr, creator, d)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(p)
	},
}

var pollsVoteCmd = &cobra.Command{
	Use:   "vote <id>",
	Short: "Vote on a poll",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		voter, _ := cmd.Flags().GetString("voter")
		opt, _ := cmd.Flags().GetInt("option")
		ctrl := PollController{}
		return ctrl.Vote(args[0], voter, opt)
	},
}

var pollsCloseCmd = &cobra.Command{Use: "close <id>", Short: "Close a poll", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	ctrl := PollController{}
	return ctrl.Close(args[0])
}}

var pollsGetCmd = &cobra.Command{Use: "get <id>", Short: "Show a poll", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	ctrl := PollController{}
	p, err := ctrl.Get(args[0])
	if err != nil {
		return err
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}}

var pollsListCmd = &cobra.Command{Use: "list", Short: "List polls", RunE: func(cmd *cobra.Command, args []string) error {
	ctrl := PollController{}
	polls, err := ctrl.List()
	if err != nil {
		return err
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(polls)
}}

func init() {
	pollsCreateCmd.Flags().String("question", "", "poll question")
	pollsCreateCmd.Flags().StringSlice("option", nil, "poll option (repeat)")
	pollsCreateCmd.Flags().String("creator", "", "creator address")
	pollsCreateCmd.Flags().String("duration", "72h", "voting duration")
	pollsCreateCmd.MarkFlagRequired("question")
	pollsCreateCmd.MarkFlagRequired("option")
	pollsCreateCmd.MarkFlagRequired("creator")

	pollsVoteCmd.Flags().String("voter", "", "voter address")
	pollsVoteCmd.Flags().Int("option", 0, "option index")
	pollsVoteCmd.MarkFlagRequired("voter")
	pollsVoteCmd.MarkFlagRequired("option")

	pollsCmd.AddCommand(pollsCreateCmd, pollsVoteCmd, pollsCloseCmd, pollsGetCmd, pollsListCmd)
}

var PollsCmd = pollsCmd

func RegisterPolls(root *cobra.Command) { root.AddCommand(PollsCmd) }
