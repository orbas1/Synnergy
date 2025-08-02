package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

func dpParseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(h)
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("invalid address")
	}
	copy(a[:], b)
	return a, nil
}

var daoProposalCmd = &cobra.Command{
	Use:   "proposal",
	Short: "Manage DAO proposals",
}

var daoProposalCreateCmd = &cobra.Command{
	Use:   "create <dao-id> <creator> <description>",
	Short: "Create a proposal",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		dur, _ := cmd.Flags().GetDuration("duration")
		creator, err := dpParseAddr(args[1])
		if err != nil {
			return err
		}
		p, err := core.CreateDAOProposal(args[0], creator, args[2], dur)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(p)
	},
}

var daoProposalVoteCmd = &cobra.Command{
	Use:   "vote <proposal-id> <voter>",
	Short: "Vote on a proposal",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokens, _ := cmd.Flags().GetUint64("tokens")
		approve, _ := cmd.Flags().GetBool("approve")
		voter, err := dpParseAddr(args[1])
		if err != nil {
			return err
		}
		return core.VoteDAOProposal(args[0], voter, tokens, approve)
	},
}

var daoProposalTallyCmd = &cobra.Command{
	Use:   "tally <proposal-id>",
	Short: "Show vote totals",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		forW, againstW, err := core.TallyDAOProposal(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "for=%d against=%d\n", forW, againstW)
		return nil
	},
}

func init() {
	daoProposalCreateCmd.Flags().Duration("duration", time.Hour, "voting period")
	daoProposalVoteCmd.Flags().Uint64("tokens", 0, "token amount to weight the vote")
	daoProposalVoteCmd.Flags().Bool("approve", true, "approve or reject")
	daoProposalVoteCmd.MarkFlagRequired("tokens")
	daoCmd.AddCommand(daoProposalCmd)
	daoProposalCmd.AddCommand(daoProposalCreateCmd, daoProposalVoteCmd, daoProposalTallyCmd)
}
