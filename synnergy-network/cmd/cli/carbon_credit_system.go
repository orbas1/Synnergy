package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

func ensureCarbon(cmd *cobra.Command, _ []string) error {
	if core.Carbon() != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	core.InitCarbonEngine(led)
	return nil
}

func mustHexCarbon(addr string) core.Address {
	return mustHex(addr)
}

var carbonCmd = &cobra.Command{
	Use:               "carbon",
	Short:             "Carbon credit project management",
	PersistentPreRunE: ensureCarbon,
}

var carbonRegisterCmd = &cobra.Command{
	Use:   "register <owner> <name> <total>",
	Short: "Register a new carbon project",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := mustHexCarbon(args[0])
		name := args[1]
		total, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("total uint64: %w", err)
		}
		id, err := core.Carbon().RegisterProject(owner, name, total)
		if err != nil {
			return err
		}
		fmt.Printf("project %d registered\n", id)
		return nil
	},
}

var carbonIssueCmd = &cobra.Command{
	Use:   "issue <projectID> <to> <amount>",
	Short: "Issue credits from a project",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("id uint64: %w", err)
		}
		to := mustHexCarbon(args[1])
		amt, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("amount uint64: %w", err)
		}
		return core.Carbon().IssueCredits(pid, to, amt)
	},
}

var carbonRetireCmd = &cobra.Command{
	Use:   "retire <holder> <amount>",
	Short: "Retire (burn) carbon credits",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		holder := mustHexCarbon(args[0])
		amt, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("amount uint64: %w", err)
		}
		return core.Carbon().RetireCredits(holder, amt)
	},
}

var carbonInfoCmd = &cobra.Command{
	Use:   "info <projectID>",
	Short: "Show project info",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("id uint64: %w", err)
		}
		proj, ok := core.Carbon().ProjectInfo(pid)
		if !ok {
			return fmt.Errorf("project not found")
		}
		b, _ := json.MarshalIndent(proj, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var carbonListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all carbon projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := core.Carbon().ListProjects()
		if err != nil {
			return err
		}
		b, _ := json.MarshalIndent(list, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

func init() {
	carbonCmd.AddCommand(carbonRegisterCmd, carbonIssueCmd, carbonRetireCmd, carbonInfoCmd, carbonListCmd)
}

var CarbonCmd = carbonCmd
