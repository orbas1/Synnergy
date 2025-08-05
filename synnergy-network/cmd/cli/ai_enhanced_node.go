package cli

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core"
)

var (
	aiNode *core.AIEnhancedNode
)

func aiNodeInit(cmd *cobra.Command, _ []string) error {
	if aiNode != nil {
		return nil
	}
	_ = godotenv.Load()

	netCfg := core.Config{
		ListenAddr:     viper.GetString("network.listen_addr"),
		BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
		DiscoveryTag:   viper.GetString("network.discovery_tag"),
	}
	ledCfg := core.LedgerConfig{WALPath: "./ledger.wal", SnapshotPath: "./ledger.snap", SnapshotInterval: 100}

	node, err := core.NewAIEnhancedNode(&core.AIEnhancedConfig{Network: netCfg, Ledger: ledCfg})
	if err != nil {
		return err
	}
	aiNode = node
	return nil
}

func aiNodeStart(cmd *cobra.Command, _ []string) error {
	if aiNode == nil {
		return fmt.Errorf("not initialised")
	}
	aiNode.Start()
	fmt.Fprintln(cmd.OutOrStdout(), "ai node started")
	return nil
}

func aiNodeStop(cmd *cobra.Command, _ []string) error {
	if aiNode == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = aiNode.Stop()
	aiNode = nil
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func aiPredictLoad(cmd *cobra.Command, args []string) error {
	if aiNode == nil {
		return fmt.Errorf("not running")
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	count, err := aiNode.PredictLoad(data)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%d\n", count)
	return nil
}

var aiNodeCmd = &cobra.Command{Use: "ainode", Short: "AI enhanced node", PersistentPreRunE: aiNodeInit}
var aiNodeStartCmd = &cobra.Command{Use: "start", Short: "Start AI node", RunE: aiNodeStart}
var aiNodeStopCmd = &cobra.Command{Use: "stop", Short: "Stop AI node", RunE: aiNodeStop}
var aiNodePredictCmd = &cobra.Command{
	Use:   "predict <file>",
	Short: "Predict tx volume",
	Args:  cobra.ExactArgs(1),
	RunE:  aiPredictLoad,
}

func init() {
	aiNodeCmd.AddCommand(aiNodeStartCmd, aiNodeStopCmd, aiNodePredictCmd)
}

// AINodeCmd exposes the root command for registration in cli/index.go
var AINodeCmd = aiNodeCmd
