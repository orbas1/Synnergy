package cli

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var tvOnce sync.Once

func tvMiddleware(cmd *cobra.Command, args []string) error {
	tvOnce.Do(func() {})
	return nil
}

// cast token vote -------------------------------------------------------------
var tokenCastCmd = &cobra.Command{
	Use:   "cast <proposal-id> <voter> <token-id> <amount> [approve]",
	Short: "Cast a token weighted vote",
	Args:  cobra.RangeArgs(4, 5),
	RunE: func(cmd *cobra.Command, args []string) error {
		voter, err := core.ParseAddress(args[1])
		if err != nil {
			return err
		}
		tid, err := strconv.ParseUint(args[2], 0, 32)
		if err != nil {
			return fmt.Errorf("token-id: %w", err)
		}
		amt, err := strconv.ParseUint(args[3], 10, 64)
		if err != nil {
			return fmt.Errorf("amount uint64: %w", err)
		}
		approve := true
		if len(args) == 5 {
			approve, err = strconv.ParseBool(args[4])
			if err != nil {
				return fmt.Errorf("approve bool: %w", err)
			}
		}
		tv := &core.TokenVote{
			ProposalID: args[0],
			Voter:      voter,
			TokenID:    core.TokenID(tid),
			Amount:     amt,
			Approve:    approve,
		}
		return core.CastTokenVote(tv)
	},
}

var TokenVoteCmd = &cobra.Command{
	Use:               "token_vote",
	Short:             "Governance voting with tokens",
	PersistentPreRunE: tvMiddleware,
}

func init() {
	TokenVoteCmd.AddCommand(tokenCastCmd)
}

func NewTokenVoteCommand() *cobra.Command { return TokenVoteCmd }
