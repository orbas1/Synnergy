package cli

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core"
)

var (
	idxNode *core.IndexingNode
	idxMu   sync.RWMutex
)

func idxInit(cmd *cobra.Command, _ []string) error {
	if idxNode != nil {
		return nil
	}
	cfg := core.IndexingNodeConfig{
		Network: core.Config{
			ListenAddr:     viper.GetString("network.listen_addr"),
			BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
			DiscoveryTag:   viper.GetString("network.discovery_tag"),
		},
	}
	led, err := core.NewLedger(core.LedgerConfig{})
	if err != nil {
		return err
	}
	node, err := core.NewIndexingNode(cfg, led)
	if err != nil {
		return err
	}
	idxMu.Lock()
	idxNode = node
	idxMu.Unlock()
	return nil
}

func idxStart(cmd *cobra.Command, _ []string) error {
	idxMu.RLock()
	n := idxNode
	idxMu.RUnlock()
	if n == nil {
		return fmt.Errorf("indexing node not initialised")
	}
	go n.ListenAndServe()
	fmt.Fprintln(cmd.OutOrStdout(), "indexing node started")
	return nil
}

func idxRebuild(cmd *cobra.Command, _ []string) error {
	idxMu.RLock()
	n := idxNode
	idxMu.RUnlock()
	if n == nil {
		return fmt.Errorf("indexing node not initialised")
	}
	for _, b := range n.Ledger().Blocks {
		n.AddBlock(b)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "index rebuilt")
	return nil
}

func idxQuery(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: query <hexAddress>")
	}
	addrBytes, err := hex.DecodeString(args[0])
	if err != nil || len(addrBytes) != 20 {
		return fmt.Errorf("invalid address")
	}
	var addr core.Address
	copy(addr[:], addrBytes)
	idxMu.RLock()
	n := idxNode
	idxMu.RUnlock()
	if n == nil {
		return fmt.Errorf("indexing node not initialised")
	}
	txs := n.QueryTxHistory(addr)
	for _, tx := range txs {
		fmt.Fprintf(cmd.OutOrStdout(), "%x -> %x : %d\n", tx.From, tx.To, tx.Amount)
	}
	return nil
}

var idxRootCmd = &cobra.Command{Use: "indexing", Short: "Run indexing node", PersistentPreRunE: idxInit}
var idxStartCmd = &cobra.Command{Use: "start", Short: "Start indexing node", RunE: idxStart}
var idxRebuildCmd = &cobra.Command{Use: "rebuild", Short: "Rebuild index from ledger", RunE: idxRebuild}
var idxQueryCmd = &cobra.Command{Use: "query", Short: "Query tx history", RunE: idxQuery}

func init() { idxRootCmd.AddCommand(idxStartCmd, idxRebuildCmd, idxQueryCmd) }

var IndexingCmd = idxRootCmd
