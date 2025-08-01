package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	core "synnergy-network/core"
)

// devnetStart launches an in-memory developer network.
func devnetStart(cmd *cobra.Command, args []string) error {
	nodes := 3
	if len(args) == 1 {
		n, err := strconv.Atoi(args[0])
		if err != nil || n <= 0 {
			return fmt.Errorf("invalid node count: %s", args[0])
		}
		nodes = n
	}
	list, err := core.StartDevNet(nodes)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "devnet started with %d nodes\n", len(list))
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	for _, n := range list {
		_ = n.Close()
	}
	return nil
}

// testnetStart spins up nodes using a YAML configuration.
func testnetStart(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: testnet start <config.yaml>")
	}
	b, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	var cfg struct {
		Nodes []core.Config `yaml:"nodes"`
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return err
	}
	list, err := core.StartTestNet(cfg.Nodes)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "testnet started with %d nodes\n", len(list))
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	for _, n := range list {
		_ = n.Close()
	}
	return nil
}

var devnetCmd = &cobra.Command{Use: "devnet", Short: "local developer network"}
var devnetStartCmd = &cobra.Command{
	Use:   "start [nodes]",
	Short: "launch N devnet nodes",
	Args:  cobra.RangeArgs(0, 1),
	RunE:  devnetStart,
}

var testnetCmd = &cobra.Command{Use: "testnet", Short: "ephemeral test network"}
var testnetStartCmd = &cobra.Command{
	Use:   "start <config.yaml>",
	Short: "start nodes from config",
	Args:  cobra.ExactArgs(1),
	RunE:  testnetStart,
}

func init() {
	devnetCmd.AddCommand(devnetStartCmd)
	testnetCmd.AddCommand(testnetStartCmd)
}

var DevnetCmd = devnetCmd
var TestnetCmd = testnetCmd
