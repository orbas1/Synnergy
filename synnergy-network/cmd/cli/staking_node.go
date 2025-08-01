package cli

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	stakeNode     *core.StakingNode
	stakeNodeOnce sync.Once
)

func stakeNodeInit(cmd *cobra.Command, _ []string) error {
	var err error
	stakeNodeOnce.Do(func() {
		_ = godotenv.Load()
		wal := envOr("LEDGER_WAL", "./ledger.wal")
		snap := envOr("LEDGER_SNAPSHOT", "./ledger.snap")
		interval := envOrInt("LEDGER_SNAPSHOT_INTERVAL", 100)
		netCfg := core.Config{
			ListenAddr:     envOr("LISTEN_ADDR", "/ip4/0.0.0.0/tcp/0"),
			BootstrapPeers: []string{},
			DiscoveryTag:   "synthron-stake",
		}
		ledCfg := core.LedgerConfig{WALPath: wal, SnapshotPath: snap, SnapshotInterval: interval}
		stakeNode, err = core.NewStakingNode(&core.StakingConfig{Network: netCfg, Ledger: ledCfg})
		if err != nil {
			return
		}
	})
	return err
}

var stakeNodeCmd = &cobra.Command{Use: "stakingnode", Short: "run staking node", PersistentPreRunE: stakeNodeInit}

var stakeNodeStartCmd = &cobra.Command{Use: "start", Short: "start staking node", RunE: func(cmd *cobra.Command, args []string) error {
	stakeNode.Start()
	fmt.Fprintln(cmd.OutOrStdout(), "staking node started")
	return nil
}}

var stakeNodeStopCmd = &cobra.Command{Use: "stop", Short: "stop staking node", RunE: func(cmd *cobra.Command, args []string) error {
	return stakeNode.Stop()
}}

var stakeNodeStakeCmd = &cobra.Command{Use: "stake <addr> <amt>", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
	addr, err := core.StringToAddress(args[0])
	if err != nil {
		return err
	}
	amt, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	return stakeNode.Stake(addr, amt)
}}

var stakeNodeUnstakeCmd = &cobra.Command{Use: "unstake <addr> <amt>", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
	addr, err := core.StringToAddress(args[0])
	if err != nil {
		return err
	}
	amt, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	return stakeNode.Unstake(addr, amt)
}}

var stakeNodeStatusCmd = &cobra.Command{Use: "status", Short: "status", RunE: func(cmd *cobra.Command, args []string) error {
	fmt.Fprintln(cmd.OutOrStdout(), stakeNode.Status())
	return nil
}}

func init() {
	stakeNodeCmd.AddCommand(stakeNodeStartCmd, stakeNodeStopCmd, stakeNodeStakeCmd, stakeNodeUnstakeCmd, stakeNodeStatusCmd)
}

var StakingNodeCmd = stakeNodeCmd

func RegisterStakingNode(root *cobra.Command) { root.AddCommand(StakingNodeCmd) }

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envOrInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
