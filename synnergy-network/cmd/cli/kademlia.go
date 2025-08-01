package cli

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"

	"synnergy-network/core"
)

var (
	kadDHT  *core.Kademlia
	kadOnce sync.Once
)

func kadInit(cmd *cobra.Command, _ []string) error {
	var err error
	kadOnce.Do(func() {
		id, _ := cmd.Flags().GetString("id")
		if id == "" {
			id = "local"
		}
		kadDHT = core.NewKademlia(core.NodeID(id))
	})
	return err
}

func kadStore(cmd *cobra.Command, args []string) error {
	kadDHT.Store(args[0], []byte(args[1]))
	fmt.Fprintln(cmd.OutOrStdout(), "stored")
	return nil
}

func kadGet(cmd *cobra.Command, args []string) error {
	val, ok := kadDHT.Lookup(args[0])
	if !ok {
		fmt.Fprintln(cmd.OutOrStdout(), "not found")
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(val))
	return nil
}

func kadAddPeer(cmd *cobra.Command, args []string) error {
	kadDHT.AddPeer(core.NodeID(args[0]))
	fmt.Fprintln(cmd.OutOrStdout(), "peer added")
	return nil
}

var kademliaCmd = &cobra.Command{
	Use:               "kademlia",
	Short:             "lightweight Kademlia DHT",
	PersistentPreRunE: kadInit,
}

var kadStoreCmd = &cobra.Command{Use: "store <key> <value>", Args: cobra.ExactArgs(2), RunE: kadStore}
var kadGetCmd = &cobra.Command{Use: "get <key>", Args: cobra.ExactArgs(1), RunE: kadGet}
var kadAddPeerCmd = &cobra.Command{Use: "addpeer <id>", Args: cobra.ExactArgs(1), RunE: kadAddPeer}

func init() {
	kademliaCmd.PersistentFlags().String("id", "local", "node identifier")
	kademliaCmd.AddCommand(kadStoreCmd, kadGetCmd, kadAddPeerCmd)
}

var KademliaCmd = kademliaCmd

func RegisterKademlia(root *cobra.Command) { root.AddCommand(KademliaCmd) }
