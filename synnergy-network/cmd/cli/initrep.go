package cli

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	logrus "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"synnergy-network/core"
)

type dummyPM struct{}

func (dummyPM) Peers() []core.PeerInfo                       { return nil }
func (dummyPM) Connect(string) error                         { return nil }
func (dummyPM) Disconnect(core.NodeID) error                 { return nil }
func (dummyPM) Sample(int) []string                          { return nil }
func (dummyPM) SendAsync(string, string, byte, []byte) error { return nil }
func (dummyPM) Subscribe(string) <-chan core.InboundMsg      { return make(chan core.InboundMsg) }
func (dummyPM) Unsubscribe(string)                           {}

var (
	initSvc     *core.InitService
	initRepOnce sync.Once
)

func initRepMiddleware(cmd *cobra.Command, _ []string) error {
	var err error
	initRepOnce.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		if e := core.InitLedger(path); e != nil {
			err = e
			return
		}
		repCfg := &core.ReplicationConfig{Fanout: 1, RequestTimeout: 5 * time.Second, SyncBatchSize: 64}
		initSvc = core.NewInitService(repCfg, logrus.StandardLogger(), core.CurrentLedger(), dummyPM{}, nil)
	})
	return err
}

func initRepStart(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()
	return initSvc.Start(ctx)
}

func initRepStop(cmd *cobra.Command, _ []string) error {
	initSvc.Shutdown()
	return nil
}

var initRepCmd = &cobra.Command{Use: "initrep", Short: "Ledger bootstrap via replication", PersistentPreRunE: initRepMiddleware}
var initRepStartCmd = &cobra.Command{Use: "start", Short: "Bootstrap and start replication", RunE: initRepStart}
var initRepStopCmd = &cobra.Command{Use: "stop", Short: "Stop replication service", RunE: initRepStop}

func init() { initRepCmd.AddCommand(initRepStartCmd, initRepStopCmd) }

var InitRepCmd = initRepCmd
