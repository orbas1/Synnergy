package cli

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var (
	bioNode *core.BiometricSecurityNode
	bioMu   sync.RWMutex
)

func bioNodeInit(cmd *cobra.Command, _ []string) error {
	if bioNode != nil {
		return nil
	}
	netCfg := core.Config{ListenAddr: ":3000"}
	ledCfg := core.LedgerConfig{WALPath: "./bio_ledger.wal", SnapshotPath: "./bio_ledger.snap"}
	n, err := core.NewBiometricSecurityNode(netCfg, ledCfg)
	if err != nil {
		return err
	}
	bioMu.Lock()
	bioNode = n
	bioMu.Unlock()
	return nil
}

func bioNodeStart(cmd *cobra.Command, _ []string) error {
	bioMu.RLock()
	n := bioNode
	bioMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not initialised")
	}
	go n.ListenAndServe()
	fmt.Fprintln(cmd.OutOrStdout(), "biometric node started")
	return nil
}

func bioNodeStop(cmd *cobra.Command, _ []string) error {
	bioMu.RLock()
	n := bioNode
	bioMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	_ = n.Close()
	bioMu.Lock()
	bioNode = nil
	bioMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func bioEnrollNode(cmd *cobra.Command, args []string) error {
	addr, _ := cmd.Flags().GetString("address")
	file, _ := cmd.Flags().GetString("file")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	bioMu.RLock()
	n := bioNode
	bioMu.RUnlock()
	if n == nil {
		return fmt.Errorf("node not running")
	}
	a, err := core.ParseAddress(addr)
	if err != nil {
		return err
	}
	return n.Enroll(a, data)
}

var bioNodeCmd = &cobra.Command{
	Use:               "bionode",
	Short:             "Run biometric security node",
	PersistentPreRunE: bioNodeInit,
}

var bioNodeStartCmd = &cobra.Command{Use: "start", Short: "start node", RunE: bioNodeStart}
var bioNodeStopCmd = &cobra.Command{Use: "stop", Short: "stop node", RunE: bioNodeStop}
var bioEnrollCmd = &cobra.Command{Use: "enroll", Short: "enroll biometric", RunE: bioEnrollNode}

func init() {
	bioEnrollCmd.Flags().String("address", "", "address")
	bioEnrollCmd.Flags().String("file", "", "data file")
	bioEnrollCmd.MarkFlagRequired("address")
	bioEnrollCmd.MarkFlagRequired("file")
	bioNodeCmd.AddCommand(bioNodeStartCmd, bioNodeStopCmd, bioEnrollCmd)
}

var BioNodeCmd = bioNodeCmd

func RegisterBioNode(root *cobra.Command) { root.AddCommand(BioNodeCmd) }
