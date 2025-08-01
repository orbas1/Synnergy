package cli

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	core "synnergy-network/core"
)

var csn *core.ConsensusSpecificNode

func csnInit(cmd *cobra.Command, _ []string) error {
	if csn != nil {
		return nil
	}
	cfg := core.Config{
		ListenAddr:     viper.GetString("network.listen_addr"),
		BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
		DiscoveryTag:   viper.GetString("network.discovery_tag"),
	}
	led, err := core.NewLedger(core.LedgerConfig{WALPath: "./ledger.wal", SnapshotPath: "./ledger.snap"})
	if err != nil {
		return err
	}
	csn, err = core.NewConsensusSpecificNode(cfg, led, nil, logrus.New())
	return err
}

func csnStart(cmd *cobra.Command, _ []string) error {
	return csn.StartConsensus()
}

func csnStop(cmd *cobra.Command, _ []string) error {
	return csn.StopConsensus()
}

var csnRootCmd = &cobra.Command{
	Use:               "consensusnode",
	Short:             "Operate a consensus specific node",
	PersistentPreRunE: csnInit,
}

var csnStartCmd = &cobra.Command{Use: "start", RunE: csnStart}
var csnStopCmd = &cobra.Command{Use: "stop", RunE: csnStop}

func init() {
	csnRootCmd.AddCommand(csnStartCmd, csnStopCmd)
}

var ConsensusNodeCmd = csnRootCmd
