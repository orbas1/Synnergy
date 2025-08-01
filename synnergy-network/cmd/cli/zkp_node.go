package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
	Nodes "synnergy-network/core/Nodes"
)

var zkpNode *Nodes.ZKPNode

func initZKP(cmd *cobra.Command, _ []string) error {
	if zkpNode != nil {
		return nil
	}
	cfg := Nodes.ZKPNodeConfig{
		Network: core.Config{ListenAddr: ":30345"},
		Ledger:  core.LedgerConfig{},
	}
	node, err := Nodes.NewZKPNode(&cfg)
	if err != nil {
		return err
	}
	zkpNode = node
	return nil
}

func zkpStart(cmd *cobra.Command, _ []string) error {
	if zkpNode == nil {
		return fmt.Errorf("not initialised")
	}
	zkpNode.Start()
	fmt.Fprintln(cmd.OutOrStdout(), "zkp node started")
	return nil
}

func zkpStop(cmd *cobra.Command, _ []string) error {
	if zkpNode == nil {
		return fmt.Errorf("not running")
	}
	_ = zkpNode.Stop()
	zkpNode = nil
	fmt.Fprintln(cmd.OutOrStdout(), "zkp node stopped")
	return nil
}

var zkpRootCmd = &cobra.Command{
	Use:               "zkpnode",
	Short:             "Run zero-knowledge proof node",
	PersistentPreRunE: initZKP,
}

var zkpStartCmd = &cobra.Command{Use: "start", Short: "Start node", RunE: zkpStart}
var zkpStopCmd = &cobra.Command{Use: "stop", Short: "Stop node", RunE: zkpStop}

func init() { zkpRootCmd.AddCommand(zkpStartCmd, zkpStopCmd) }

var ZKPNodeCmd = zkpRootCmd
