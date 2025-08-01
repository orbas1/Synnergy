package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

func ensurePension(cmd *cobra.Command, _ []string) error {
	if core.Pension() != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	core.InitPensionEngine(led)
	return nil
}

func parseScheduleArg(arg string) ([]core.VestingEntry, error) {
	if arg == "" {
		return nil, nil
	}
	parts := strings.Split(arg, ",")
	var out []core.VestingEntry
	for _, p := range parts {
		sp := strings.Split(p, ":")
		if len(sp) != 2 {
			return nil, fmt.Errorf("bad schedule point %s", p)
		}
		ts, err := strconv.ParseInt(sp[0], 10, 64)
		if err != nil {
			return nil, err
		}
		amt, err := strconv.ParseUint(sp[1], 10, 64)
		if err != nil {
			return nil, err
		}
		out = append(out, core.VestingEntry{Timestamp: ts, Amount: amt})
	}
	return out, nil
}

var pensionCmd = &cobra.Command{
	Use:               "pension",
	Short:             "Pension fund management",
	PersistentPreRunE: ensurePension,
}

var pensionRegisterCmd = &cobra.Command{
	Use:   "register <owner> <name> <maturity> [schedule]",
	Short: "Register a new pension plan",
	Args:  cobra.RangeArgs(3, 4),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := mustHex(args[0])
		name := args[1]
		mat, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("maturity int64: %w", err)
		}
		var sched []core.VestingEntry
		if len(args) == 4 {
			sched, err = parseScheduleArg(args[3])
			if err != nil {
				return err
			}
		}
		id, err := core.Pension().RegisterPlan(owner, name, time.Unix(mat, 0), sched)
		if err != nil {
			return err
		}
		fmt.Printf("plan %d registered\n", id)
		return nil
	},
}

var pensionContribCmd = &cobra.Command{
	Use:   "contribute <planID> <to> <amount>",
	Short: "Contribute to a plan",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("id uint64: %w", err)
		}
		to := mustHex(args[1])
		amt, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("amount uint64: %w", err)
		}
		return core.Pension().Contribute(pid, to, amt)
	},
}

var pensionWithdrawCmd = &cobra.Command{
	Use:   "withdraw <planID> <holder> <amount>",
	Short: "Withdraw vested tokens",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("id uint64: %w", err)
		}
		holder := mustHex(args[1])
		amt, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("amount uint64: %w", err)
		}
		return core.Pension().Withdraw(pid, holder, amt)
	},
}

var pensionInfoCmd = &cobra.Command{
	Use:   "info <planID>",
	Short: "Show plan info",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("id uint64: %w", err)
		}
		p, ok := core.Pension().PlanInfo(pid)
		if !ok {
			return fmt.Errorf("plan not found")
		}
		b, _ := json.MarshalIndent(p, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var pensionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pension plans",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := core.Pension().ListPlans()
		if err != nil {
			return err
		}
		b, _ := json.MarshalIndent(list, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

func init() {
	pensionCmd.AddCommand(pensionRegisterCmd, pensionContribCmd, pensionWithdrawCmd, pensionInfoCmd, pensionListCmd)
}

var PensionCmd = pensionCmd
