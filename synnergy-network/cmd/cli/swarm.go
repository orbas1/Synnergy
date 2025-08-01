package cli

// swarm.go - manage groups of network nodes

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"synnergy-network/core"
)

var (
	swarmOnce sync.Once
	swarmCtl  *core.Swarm
)

func swarmInit(cmd *cobra.Command, _ []string) error {
	var err error
	swarmOnce.Do(func() {
		path := viper.GetString("ledger.path")
		if path == "" {
			path = os.Getenv("LEDGER_PATH")
		}
		var led *core.Ledger
		if path != "" {
			led, err = core.OpenLedger(path)
			if err != nil {
				return
			}
		}
		swarmCtl = core.NewSwarm(led, nil)
	})
	return err
}

func swarmAdd(cmd *cobra.Command, args []string) error {
	cfg := core.Config{ListenAddr: args[1]}
	n, err := core.NewNode(cfg)
	if err != nil {
		return err
	}
	return swarmCtl.AddNode(core.NodeID(args[0]), n)
}

func swarmRemove(cmd *cobra.Command, args []string) error {
	swarmCtl.RemoveNode(core.NodeID(args[0]))
	return nil
}

func swarmBroadcast(cmd *cobra.Command, args []string) error {
	b, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	var tx core.Transaction
	if err := json.Unmarshal(b, &tx); err != nil {
		return err
	}
	return swarmCtl.BroadcastTx(&tx)
}

func swarmPeers(cmd *cobra.Command, _ []string) error {
	for _, id := range swarmCtl.Peers() {
		fmt.Fprintln(cmd.OutOrStdout(), id)
	}
	return nil
}

func swarmStart(cmd *cobra.Command, _ []string) error {
	swarmCtl.Start(cmd.Context())
	return nil
}

func swarmStop(cmd *cobra.Command, _ []string) error {
	swarmCtl.Stop()
	return nil
}

var swarmCmd = &cobra.Command{Use: "swarm", Short: "Manage node swarms", PersistentPreRunE: swarmInit}

var swarmAddCmd = &cobra.Command{Use: "add <id> <addr>", Args: cobra.ExactArgs(2), RunE: swarmAdd}
var swarmRemoveCmd = &cobra.Command{Use: "remove <id>", Args: cobra.ExactArgs(1), RunE: swarmRemove}
var swarmBroadcastCmd = &cobra.Command{Use: "broadcast <tx.json>", Args: cobra.ExactArgs(1), RunE: swarmBroadcast}
var swarmPeersCmd = &cobra.Command{Use: "peers", Args: cobra.NoArgs, RunE: swarmPeers}
var swarmStartCmd = &cobra.Command{Use: "start", Args: cobra.NoArgs, RunE: swarmStart}
var swarmStopCmd = &cobra.Command{Use: "stop", Args: cobra.NoArgs, RunE: swarmStop}

func init() {
	swarmCmd.AddCommand(swarmAddCmd, swarmRemoveCmd, swarmBroadcastCmd, swarmPeersCmd, swarmStartCmd, swarmStopCmd)
}

var SwarmCmd = swarmCmd
