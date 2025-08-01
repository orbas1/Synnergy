package cli

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
	Nodes "synnergy-network/core/Nodes"
)

var electedNode *core.ElectedAuthorityNode

func ensureElectedNode(cmd *cobra.Command, _ []string) error {
	if electedNode != nil {
		return nil
	}
	// in CLI prototypes we create an in-memory node without networking
	electedNode = core.NewElectedAuthorityNode(nil, core.CurrentLedger(), nil, 1)
	return nil
}

type ElectedAuthController struct{}

func (c *ElectedAuthController) vote(addr string) {
	var a core.Address
	b, _ := hex.DecodeString(strings.TrimPrefix(addr, "0x"))
	copy(a[:], b)
	var na Nodes.Address
	copy(na[:], a[:])
	electedNode.RecordVote(na)
}

func (c *ElectedAuthController) report(addr string) {
	var a core.Address
	b, _ := hex.DecodeString(strings.TrimPrefix(addr, "0x"))
	copy(a[:], b)
	var na Nodes.Address
	copy(na[:], a[:])
	electedNode.ReportMisbehaviour(na)
}

//---------------------------------------------------------------------
// CLI commands
//---------------------------------------------------------------------

var electedAuthCmd = &cobra.Command{
	Use:               "electedauth",
	Short:             "Manage elected authority node",
	PersistentPreRunE: ensureElectedNode,
}

var electedVoteCmd = &cobra.Command{
	Use:   "vote <addr>",
	Short: "Vote to elect the node",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctrl := &ElectedAuthController{}
		ctrl.vote(args[0])
		fmt.Fprintln(cmd.OutOrStdout(), "vote recorded")
	},
}

var electedReportCmd = &cobra.Command{
	Use:   "report <addr>",
	Short: "Report the node for misbehaviour",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctrl := &ElectedAuthController{}
		ctrl.report(args[0])
		fmt.Fprintln(cmd.OutOrStdout(), "report recorded")
	},
}

func init() {
	electedAuthCmd.AddCommand(electedVoteCmd, electedReportCmd)
}

var ElectedAuthCmd = electedAuthCmd
