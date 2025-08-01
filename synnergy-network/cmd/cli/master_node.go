package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var master *core.MasterNode

func ensureMaster(cmd *cobra.Command, args []string) error {
	if master != nil {
		return nil
	}
	n, _ := core.NewNode(core.Config{})
	master = core.NewMasterNode(n, &core.Ledger{}, &core.SynnergyConsensus{}, nil, core.Address{}, 0)
	return nil
}

var masterCmd = &cobra.Command{
	Use:               "master",
	Short:             "Master node operations",
	PersistentPreRunE: ensureMaster,
}

var masterStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start master node services",
	RunE: func(cmd *cobra.Command, args []string) error {
		master.Start()
		fmt.Println("master node started")
		return nil
	},
}

var masterStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop master node services",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := master.Stop(); err != nil {
			return err
		}
		fmt.Println("master node stopped")
		return nil
	},
}

func init() {
	masterCmd.AddCommand(masterStartCmd)
	masterCmd.AddCommand(masterStopCmd)
}

var MasterCmd = masterCmd
