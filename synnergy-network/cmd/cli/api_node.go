package cli

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"

	"synnergy-network/core"
)

var (
	apiNode *core.APINode
	apiMu   sync.RWMutex
)

func apiInit(cmd *cobra.Command, _ []string) error {
	if apiNode != nil {
		return nil
	}
	nCfg := core.Config{ListenAddr: "/ip4/0.0.0.0/tcp/4001"}
	n, err := core.NewNode(nCfg)
	if err != nil {
		return err
	}
	led, err := core.NewLedger(core.LedgerConfig{})
	if err != nil {
		return err
	}
	apiMu.Lock()
	apiNode = core.NewAPINode(n, led)
	apiMu.Unlock()
	return nil
}

func apiStart(cmd *cobra.Command, _ []string) error {
	apiMu.RLock()
	n := apiNode
	apiMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not initialised")
	}
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		_ = n.APINode_Stop()
		os.Exit(0)
	}()
	return n.APINode_Start(cmd.Flag("addr").Value.String())
}

func apiStop(cmd *cobra.Command, _ []string) error {
	apiMu.RLock()
	n := apiNode
	apiMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	return n.APINode_Stop()
}

var apiCmd = &cobra.Command{Use: "api-node", Short: "Run API gateway", PersistentPreRunE: apiInit}
var apiStartCmd = &cobra.Command{Use: "start", Short: "Start API node", RunE: apiStart}
var apiStopCmd = &cobra.Command{Use: "stop", Short: "Stop API node", RunE: apiStop}

func init() {
	apiStartCmd.Flags().String("addr", ":8080", "listen address")
	apiCmd.AddCommand(apiStartCmd, apiStopCmd)
}

var APINodeCmd = apiCmd

func RegisterAPINode(root *cobra.Command) { root.AddCommand(APINodeCmd) }
