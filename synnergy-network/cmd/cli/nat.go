package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"synnergy-network/core"
)

var (
	natMgr *core.NATManager
)

func natInit(cmd *cobra.Command, _ []string) error {
	if natMgr != nil {
		return nil
	}
	m, err := core.NewNATManager()
	if err != nil {
		return err
	}
	natMgr = m
	return nil
}

func natMap(cmd *cobra.Command, args []string) error {
	port, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	if err := natMgr.Map(port); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "mapped ✔")
	return nil
}

func natUnmap(cmd *cobra.Command, _ []string) error {
	if err := natMgr.Unmap(); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "unmapped ✔")
	return nil
}

func natIP(cmd *cobra.Command, _ []string) error {
	fmt.Fprintln(cmd.OutOrStdout(), natMgr.ExternalIP())
	return nil
}

var natRootCmd = &cobra.Command{Use: "nat", Short: "NAT traversal", PersistentPreRunE: natInit}
var natMapCmd = &cobra.Command{Use: "map <port>", Short: "Map port", Args: cobra.ExactArgs(1), RunE: natMap}
var natUnmapCmd = &cobra.Command{Use: "unmap", Short: "Unmap port", Args: cobra.NoArgs, RunE: natUnmap}
var natIPCmd = &cobra.Command{Use: "ip", Short: "Show external IP", Args: cobra.NoArgs, RunE: natIP}

func init() { natRootCmd.AddCommand(natMapCmd, natUnmapCmd, natIPCmd) }

// NatCmd exposes the NAT traversal commands.
var NatCmd = natRootCmd

// RegisterNAT adds NAT traversal commands to the root CLI.
func RegisterNAT(root *cobra.Command) { root.AddCommand(NatCmd) }
