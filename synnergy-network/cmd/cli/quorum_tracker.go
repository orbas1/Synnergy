package cli

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"

	"synnergy-network/core"
)

var quorumCmd = &cobra.Command{
	Use:   "quorum",
	Short: "Manage quorum trackers",
}

var quorumInitCmd = &cobra.Command{
	Use:   "init [total] [threshold]",
	Short: "Initialise a quorum tracker",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		total, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		thr, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}
		core.InitQuorumTracker(total, thr)
		logrus.Infof("quorum tracker initialised: %d/%d", thr, total)
		return nil
	},
}

var quorumVoteCmd = &cobra.Command{
	Use:   "vote [address]",
	Short: "Record a vote",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := core.ParseAddress(args[0])
		if err != nil {
			return err
		}
		n := core.QuorumAddVote(addr)
		logrus.Infof("vote recorded (%d total)", n)
		return nil
	},
}

var quorumCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if quorum reached",
	Run: func(cmd *cobra.Command, args []string) {
		if core.QuorumHasQuorum() {
			logrus.Info("quorum reached")
		} else {
			logrus.Info("quorum not reached")
		}
	},
}

var quorumResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Clear recorded votes",
	Run: func(cmd *cobra.Command, args []string) {
		core.QuorumReset()
		logrus.Info("quorum tracker reset")
	},
}

func init() {
	quorumCmd.AddCommand(quorumInitCmd, quorumVoteCmd, quorumCheckCmd, quorumResetCmd)
}
