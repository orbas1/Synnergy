package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"synnergy-network/core"
)

var (
	peerMgr *core.PeerManagement
)

func peerInit(cmd *cobra.Command, _ []string) error {
	if err := netInit(cmd, nil); err != nil {
		return err
	}
	netMu.RLock()
	n := netNode
	netMu.RUnlock()
	if n == nil {
		return fmt.Errorf("network not running")
	}
	if peerMgr == nil {
		peerMgr = core.NewPeerManagement(n)
	}
	return nil
}

func peerDiscover(cmd *cobra.Command, _ []string) error {
	infos := peerMgr.DiscoverPeers()
	for _, p := range infos {
		fmt.Fprintf(cmd.OutOrStdout(), "%v\n", p)
	}
	return nil
}

func peerConnect(cmd *cobra.Command, args []string) error {
	return peerMgr.Connect(args[0])
}

func peerAdvertise(cmd *cobra.Command, args []string) error {
	topic := "synnergy-peer"
	if len(args) > 0 {
		topic = args[0]
	}
	return peerMgr.AdvertiseSelf(topic)
}

var peerCmd = &cobra.Command{Use: "peer", Short: "Peer management", PersistentPreRunE: peerInit}
var peerDiscoverCmd = &cobra.Command{Use: "discover", Short: "List known peers", RunE: peerDiscover}
var peerConnectCmd = &cobra.Command{Use: "connect <multiaddr>", Short: "Connect to peer", Args: cobra.ExactArgs(1), RunE: peerConnect}
var peerAdvertiseCmd = &cobra.Command{Use: "advertise [topic]", Short: "Advertise this node", Args: cobra.RangeArgs(0, 1), RunE: peerAdvertise}

func init() {
	peerCmd.AddCommand(peerDiscoverCmd)
	peerCmd.AddCommand(peerConnectCmd)
	peerCmd.AddCommand(peerAdvertiseCmd)
}

var PeerCmd = peerCmd
