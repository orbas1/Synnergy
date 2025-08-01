package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
	Tokens "synnergy-network/core/Tokens"
)

func ensureSYN200(cmd *cobra.Command, _ []string) error {
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

func mustAddrSYN200(a string) core.Address { return mustHex(a) }

var syn200Cmd = &cobra.Command{
	Use:               "syn200",
	Short:             "Manage SYN200 carbon credit tokens",
	PersistentPreRunE: ensureSYN200,
}

var syn200RegisterCmd = &cobra.Command{
	Use:   "register <owner> <name> <total>",
	Short: "Register a new carbon project",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := mustAddrSYN200(args[0])
		name := args[1]
		total, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("total uint64: %w", err)
		}
		id, err := core.Carbon().RegisterProject(owner, name, total)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "project %d registered\n", id)
		return nil
	},
}

var syn200IssueCmd = &cobra.Command{
	Use:   "issue <projectID> <to> <amount>",
	Short: "Issue credits from a project",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("id uint64: %w", err)
		}
		to := mustAddrSYN200(args[1])
		amt, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("amount uint64: %w", err)
		}
		return core.Carbon().IssueCredits(pid, to, amt)
	},
}

var syn200RetireCmd = &cobra.Command{
	Use:   "retire <holder> <amount>",
	Short: "Retire carbon credits",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		holder := mustAddrSYN200(args[0])
		amt, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("amount uint64: %w", err)
		}
		return core.Carbon().RetireCredits(holder, amt)
	},
}

var syn200VerifyCmd = &cobra.Command{
	Use:   "verify <projectID> <verifier> <verID> [status]",
	Short: "Add verification record",
	Args:  cobra.RangeArgs(3, 4),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("id uint64: %w", err)
		}
		ver := mustAddrSYN200(args[1])
		verID := args[2]
		status := "verified"
		if len(args) == 4 {
			status = args[3]
		}
		rec := Tokens.VerificationRecord{ID: verID, Verifier: ver, Timestamp: time.Now(), Status: status}
		return core.Carbon().AddVerification(pid, rec)
	},
}

var syn200VerListCmd = &cobra.Command{
	Use:   "verifications <projectID>",
	Short: "List verification records",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("id uint64: %w", err)
		}
		list, err := core.Carbon().ListVerifications(pid)
		if err != nil {
			return err
		}
		b, _ := json.MarshalIndent(list, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(b))
		return nil
	},
}

var syn200InfoCmd = &cobra.Command{
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
		fmt.Fprintln(cmd.OutOrStdout(), string(b))
		return nil
	},
}

var syn200ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all carbon projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := core.Carbon().ListProjects()
		if err != nil {
			return err
		}
		b, _ := json.MarshalIndent(list, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(b))
		return nil
	},
}

func init() {
	syn200Cmd.AddCommand(syn200RegisterCmd, syn200IssueCmd, syn200RetireCmd, syn200VerifyCmd, syn200VerListCmd, syn200InfoCmd, syn200ListCmd)
}

var SYN200Cmd = syn200Cmd
