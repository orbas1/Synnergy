package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core"
)

var (
	envNode *core.EnvironmentalMonitoringNode
	envMu   sync.RWMutex
)

func envInit(cmd *cobra.Command, _ []string) error {
	if envNode != nil {
		return nil
	}
	_ = godotenv.Load()

	lv, err := logrus.ParseLevel(viper.GetString("logging.level"))
	if err != nil {
		return err
	}
	logrus.SetLevel(lv)

	netCfg := core.Config{
		ListenAddr:     viper.GetString("network.listen_addr"),
		BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
		DiscoveryTag:   viper.GetString("network.discovery_tag"),
	}
	path := viper.GetString("ledger.path")
	if path == "" {
		path = os.Getenv("LEDGER_PATH")
	}
	var led *core.Ledger
	if path != "" {
		led, err = core.OpenLedger(path)
		if err != nil {
			return err
		}
	} else {
		led, err = core.NewLedger(core.LedgerConfig{})
		if err != nil {
			return err
		}
	}
	node, err := core.NewEnvironmentalMonitoringNode(netCfg, led)
	if err != nil {
		return err
	}
	envMu.Lock()
	envNode = node
	envMu.Unlock()
	return nil
}

func envStart(cmd *cobra.Command, _ []string) error {
	envMu.RLock()
	n := envNode
	envMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not initialised")
	}
	n.Start()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		_ = n.Stop()
		os.Exit(0)
	}()
	fmt.Fprintln(cmd.OutOrStdout(), "environmental node started")
	return nil
}

func envStop(cmd *cobra.Command, _ []string) error {
	envMu.RLock()
	n := envNode
	envMu.RUnlock()
	if n == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = n.Stop()
	envMu.Lock()
	envNode = nil
	envMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func envRegister(cmd *cobra.Command, args []string) error {
	envMu.RLock()
	n := envNode
	envMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	return n.RegisterSensor(args[0], args[1])
}

func envList(cmd *cobra.Command, _ []string) error {
	envMu.RLock()
	n := envNode
	envMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	sensors, err := n.ListSensors()
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(sensors)
}

var envCmd = &cobra.Command{
	Use:               "envnode",
	Short:             "Run environmental monitoring node",
	PersistentPreRunE: envInit,
}

var envStartCmd = &cobra.Command{Use: "start", RunE: envStart}
var envStopCmd = &cobra.Command{Use: "stop", RunE: envStop}
var envRegisterCmd = &cobra.Command{Use: "register [id] [endpoint]", Args: cobra.ExactArgs(2), RunE: envRegister}
var envListCmd = &cobra.Command{Use: "list", RunE: envList}

func init() {
	envCmd.AddCommand(envStartCmd, envStopCmd, envRegisterCmd, envListCmd)
}

var EnvironmentalNodeCmd = envCmd
