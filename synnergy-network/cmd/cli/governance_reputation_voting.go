package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var repGovCmd = &cobra.Command{
	Use:     "repvote",
	Short:   "Reputation weighted governance",
	Aliases: []string{"repvote", "repgov"},
}

var repProposeCmd = &cobra.Command{
	Use:   "propose",
	Short: "Submit a reputation proposal",
	RunE: func(cmd *cobra.Command, args []string) error {
		changesStr, _ := cmd.Flags().GetString("changes")
		desc, _ := cmd.Flags().GetString("desc")
		dlStr, _ := cmd.Flags().GetString("deadline")
		if changesStr == "" {
			return errors.New("--changes required")
		}
		changes := map[string]string{}
		for _, pair := range splitPairs(changesStr) {
			kv := splitKV(pair)
			if len(kv) != 2 {
				return fmt.Errorf("invalid change pair %q", pair)
			}
			changes[kv[0]] = kv[1]
		}
		p := &core.RepGovProposal{Changes: changes, Description: desc}
		if dlStr != "" {
			d, err := time.ParseDuration(dlStr)
			if err != nil {
				return err
			}
			p.Deadline = time.Now().Add(d)
		}
		if err := core.SubmitRepGovProposal(p); err != nil {
			return err
		}
		fmt.Printf("Proposal submitted: %s\n", p.ID)
		return nil
	},
}

var repVoteCmd = &cobra.Command{
	Use:   "vote [proposal-id]",
	Short: "Cast a reputation vote",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		approve, _ := cmd.Flags().GetBool("approve")
		addrStr, _ := cmd.Flags().GetString("addr")
		addr, err := core.StringToAddress(addrStr)
		if err != nil {
			return err
		}
		return core.CastRepGovVote(args[0], addr, approve)
	},
}

var repExecCmd = &cobra.Command{
	Use:   "execute [proposal-id]",
	Short: "Execute a reputation proposal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return core.ExecuteRepGovProposal(args[0])
	},
}

var repGetCmd = &cobra.Command{
	Use:   "get [proposal-id]",
	Short: "Show a reputation proposal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := core.GetRepGovProposal(args[0])
		if err != nil {
			return err
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(p)
	},
}

var repListCmd = &cobra.Command{
	Use:   "list",
	Short: "List reputation proposals",
	RunE: func(cmd *cobra.Command, args []string) error {
		props, err := core.ListRepGovProposals()
		if err != nil {
			return err
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(props)
	},
}

var repBalCmd = &cobra.Command{
	Use:   "balance [address]",
	Short: "Show reputation balance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := core.StringToAddress(args[0])
		if err != nil {
			return err
		}
		bal, err := core.ReputationOf(addr)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%d\n", bal)
		return nil
	},
}

func splitPairs(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func splitKV(s string) []string { return strings.SplitN(strings.TrimSpace(s), "=", 2) }

func init() {
	repProposeCmd.Flags().String("changes", "", "comma-separated key=value list")
	repProposeCmd.Flags().String("desc", "", "proposal description")
	repProposeCmd.Flags().String("deadline", "", "deadline duration")
	repVoteCmd.Flags().Bool("approve", true, "approve or reject")
	repVoteCmd.Flags().String("addr", "", "voter address")
	repGovCmd.AddCommand(repProposeCmd, repVoteCmd, repExecCmd, repGetCmd, repListCmd, repBalCmd)
}

func NewRepGovCommand() *cobra.Command { return repGovCmd }
