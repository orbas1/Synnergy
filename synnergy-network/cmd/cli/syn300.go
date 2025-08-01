package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var syn300Cmd = &cobra.Command{
	Use:   "syn300",
	Short: "Manage SYN300 governance token",
}

func syn300Resolve() (*core.SYN300Token, error) {
	for _, t := range core.GetRegistryTokens() {
		if t.Meta().Standard == core.StdSYN300 {
			if gt, ok := any(t).(*core.SYN300Token); ok {
				return gt, nil
			}
		}
	}
	return nil, fmt.Errorf("SYN300 token not found")
}

var syn300DelegateCmd = &cobra.Command{
	Use:   "delegate <owner> <delegate>",
	Short: "Delegate voting power",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		gt, err := syn300Resolve()
		if err != nil {
			return err
		}
		owner, err := core.ParseAddress(args[0])
		if err != nil {
			return err
		}
		del, err := core.ParseAddress(args[1])
		if err != nil {
			return err
		}
		gt.Delegate(owner, del)
		fmt.Fprintln(cmd.OutOrStdout(), "delegated")
		return nil
	},
}

var syn300RevokeCmd = &cobra.Command{
	Use:   "revoke <owner>",
	Short: "Revoke delegation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		gt, err := syn300Resolve()
		if err != nil {
			return err
		}
		owner, err := core.ParseAddress(args[0])
		if err != nil {
			return err
		}
		gt.RevokeDelegate(owner)
		fmt.Fprintln(cmd.OutOrStdout(), "revoked")
		return nil
	},
}

var syn300PowerCmd = &cobra.Command{
	Use:   "power <addr>",
	Short: "Show voting power",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		gt, err := syn300Resolve()
		if err != nil {
			return err
		}
		addr, err := core.ParseAddress(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%d\n", gt.VotingPower(addr))
		return nil
	},
}

var syn300ProposeCmd = &cobra.Command{
	Use:   "propose <creator> <description>",
	Short: "Create proposal",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		gt, err := syn300Resolve()
		if err != nil {
			return err
		}
		creator, err := core.ParseAddress(args[0])
		if err != nil {
			return err
		}
		id := gt.CreateProposal(creator, args[1], 72*time.Hour)
		fmt.Fprintf(cmd.OutOrStdout(), "%d\n", id)
		return nil
	},
}

var syn300VoteCmd = &cobra.Command{
	Use:   "vote <id> <voter> <approve>",
	Short: "Vote on proposal",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		gt, err := syn300Resolve()
		if err != nil {
			return err
		}
		id, err := parseUint64(args[0])
		if err != nil {
			return err
		}
		voter, err := core.ParseAddress(args[1])
		if err != nil {
			return err
		}
		approve, err := parseBool(args[2])
		if err != nil {
			return err
		}
		gt.Vote(id, voter, approve)
		fmt.Fprintln(cmd.OutOrStdout(), "voted")
		return nil
	},
}

var syn300ExecCmd = &cobra.Command{
	Use:   "execute <id> <quorum>",
	Short: "Execute proposal",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		gt, err := syn300Resolve()
		if err != nil {
			return err
		}
		id, err := parseUint64(args[0])
		if err != nil {
			return err
		}
		q, err := parseUint64(args[1])
		if err != nil {
			return err
		}
		ok := gt.ExecuteProposal(id, q)
		fmt.Fprintf(cmd.OutOrStdout(), "%v\n", ok)
		return nil
	},
}

var syn300StatusCmd = &cobra.Command{
	Use:   "status <id>",
	Short: "Get proposal status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		gt, err := syn300Resolve()
		if err != nil {
			return err
		}
		id, err := parseUint64(args[0])
		if err != nil {
			return err
		}
		p, ok := gt.ProposalStatus(id)
		if !ok {
			return fmt.Errorf("not found")
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%+v\n", *p)
		return nil
	},
}

var syn300ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List proposals",
	RunE: func(cmd *cobra.Command, args []string) error {
		gt, err := syn300Resolve()
		if err != nil {
			return err
		}
		for _, p := range gt.ListProposals() {
			fmt.Fprintf(cmd.OutOrStdout(), "%d\t%s\t%v\n", p.ID, p.Description, p.Executed)
		}
		return nil
	},
}

func init() {
	syn300Cmd.AddCommand(syn300DelegateCmd, syn300RevokeCmd, syn300PowerCmd, syn300ProposeCmd, syn300VoteCmd, syn300ExecCmd, syn300StatusCmd, syn300ListCmd)
}

func NewSYN300Command() *cobra.Command { return syn300Cmd }

func parseUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 0, 64)
}

func parseBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}
