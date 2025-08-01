package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"synnergy-network/core"
)

// -----------------------------------------------------------------------------
// vm_sandbox_management.go - manage VM sandboxes via CLI
// -----------------------------------------------------------------------------

var sandboxCmd = &cobra.Command{Use: "sandbox", Short: "VM sandbox management"}

var sandboxStartCmd = &cobra.Command{
	Use:   "start <contract> <mem> <cpu>",
	Short: "Start sandbox",
	Args:  cobra.ExactArgs(3),
	RunE:  sandboxHandleStart,
}

var sandboxStopCmd = &cobra.Command{
	Use:   "stop <contract>",
	Short: "Stop sandbox",
	Args:  cobra.ExactArgs(1),
	RunE:  sandboxHandleStop,
}

var sandboxResetCmd = &cobra.Command{
	Use:   "reset <contract>",
	Short: "Reset sandbox",
	Args:  cobra.ExactArgs(1),
	RunE:  sandboxHandleReset,
}

var sandboxStatusCmd = &cobra.Command{
	Use:   "status <contract>",
	Short: "Sandbox status",
	Args:  cobra.ExactArgs(1),
	RunE:  sandboxHandleStatus,
}

var sandboxListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sandboxes",
	Args:  cobra.NoArgs,
	RunE:  sandboxHandleList,
}

func init() {
	sandboxCmd.AddCommand(sandboxStartCmd, sandboxStopCmd, sandboxResetCmd, sandboxStatusCmd, sandboxListCmd)
}

func sandboxHandleStart(cmd *cobra.Command, args []string) error {
	addr, err := parseAddr(args[0])
	if err != nil {
		return err
	}
	mem, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	cpu, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return err
	}
	if err := core.StartSandbox(addr, mem, cpu); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "started")
	return nil
}

func sandboxHandleStop(cmd *cobra.Command, args []string) error {
	addr, err := parseAddr(args[0])
	if err != nil {
		return err
	}
	if err := core.StopSandbox(addr); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func sandboxHandleReset(cmd *cobra.Command, args []string) error {
	addr, err := parseAddr(args[0])
	if err != nil {
		return err
	}
	if err := core.ResetSandbox(addr); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "reset")
	return nil
}

func sandboxHandleStatus(cmd *cobra.Command, args []string) error {
	addr, err := parseAddr(args[0])
	if err != nil {
		return err
	}
	info, ok := core.SandboxStatus(addr)
	if !ok {
		fmt.Fprintln(cmd.OutOrStdout(), "not found")
		return nil
	}
	out, _ := json.MarshalIndent(info, "", "  ")
	cmd.OutOrStdout().Write(out)
	cmd.OutOrStdout().Write([]byte("\n"))
	return nil
}

func sandboxHandleList(cmd *cobra.Command, _ []string) error {
	list := core.ListSandboxes()
	out, _ := json.MarshalIndent(list, "", "  ")
	cmd.OutOrStdout().Write(out)
	cmd.OutOrStdout().Write([]byte("\n"))
	return nil
}

// SandboxCmd is the exported command for registration in index.go
var SandboxCmd = sandboxCmd

func RegisterSandbox(root *cobra.Command) { root.AddCommand(SandboxCmd) }
