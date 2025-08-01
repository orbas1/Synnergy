package cli

import (
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	core "synnergy-network/core"
	Nodes "synnergy-network/core/Nodes"
)

var (
	holo    *Nodes.HolographicNode
	holoNet core.NodeInterface
	holoLed *core.Ledger
	holoMu  sync.RWMutex
)

func holoInit(cmd *cobra.Command, _ []string) error {
	if holo != nil {
		return nil
	}
	if err := netInit(cmd, nil); err != nil {
		return err
	}
	netMu.RLock()
	holoNet = netNode
	netMu.RUnlock()
	led, err := core.OpenLedger(viper.GetString("ledger.path"))
	if err != nil {
		return err
	}
	holoLed = led
	holo = Nodes.NewHolographicNode(holoNet, holoLed)
	return nil
}

func holoStart(cmd *cobra.Command, _ []string) error {
	holo.HoloStart()
	return nil
}

func holoStop(cmd *cobra.Command, _ []string) error {
	return holo.HoloStop()
}

var holoCmd = &cobra.Command{
	Use:               "holo",
	Short:             "Manage holographic node",
	PersistentPreRunE: holoInit,
}

var holoStartCmd = &cobra.Command{Use: "start", RunE: holoStart}
var holoStopCmd = &cobra.Command{Use: "stop", RunE: holoStop}

func init() {
	holoCmd.AddCommand(holoStartCmd, holoStopCmd)
}

// HoloCmd exported for index
var HoloCmd = holoCmd

func RegisterHolo(root *cobra.Command) { root.AddCommand(HoloCmd) }
