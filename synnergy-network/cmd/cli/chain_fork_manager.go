package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var (
	forkOnce   sync.Once
	forkLedger *core.Ledger
	forkErr    error
)

func forkInit(cmd *cobra.Command, _ []string) error {
	forkOnce.Do(func() {
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			forkErr = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		forkLedger, forkErr = core.OpenLedger(path)
		if forkErr == nil {
			forkErr = core.InitForkManager(forkLedger)
		}
	})
	return forkErr
}

// controller helpers
func forkList(_ *cobra.Command, _ []string) error {
	forks := core.ListForks()
	data, _ := json.MarshalIndent(forks, "", "  ")
	fmt.Println(string(data))
	return nil
}

func forkResolve(_ *cobra.Command, _ []string) error {
	return core.ResolveForks()
}

func forkRecover(_ *cobra.Command, _ []string) error {
	return core.RecoverLongestFork()
}

var forkCmd = &cobra.Command{
	Use:               "fork",
	Short:             "Manage chain forks",
	PersistentPreRunE: forkInit,
}

var forkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List known forks",
	RunE:  forkList,
}

var forkResolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolve forks extending the current tip",
	RunE:  forkResolve,
}

var forkRecoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Rebuild the chain to the longest known fork",
	RunE:  forkRecover,
}

func init() {
	forkCmd.AddCommand(forkListCmd, forkResolveCmd, forkRecoverCmd)
}

var ForkCmd = forkCmd
