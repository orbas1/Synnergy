package cli

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	lpApply *core.LoanPoolApply
)

func ensureLoanApply(cmd *cobra.Command, _ []string) error {
	if lpApply != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return errors.New("ledger not initialised")
	}
	stdlog := log.New(logrus.StandardLogger().Out, "", log.LstdFlags)
	lpApply = core.NewLoanPoolApply(stdlog, led, time.Hour*24, 0)
	return nil
}

// Controller thin wrapper

type LoanApplyController struct{}

func (c *LoanApplyController) Submit(applicant string, amt uint64, term uint16, purpose string) (core.Hash, error) {
	a, err := core.StringToAddress(applicant)
	if err != nil {
		return core.Hash{}, err
	}
	return lpApply.Submit(a, amt, term, purpose)
}

func (c *LoanApplyController) Vote(voter string, id core.Hash, approve bool) error {
	v, err := core.StringToAddress(voter)
	if err != nil {
		return err
	}
	return lpApply.Vote(v, id, approve)
}

func (c *LoanApplyController) Process(ts time.Time)        { lpApply.Process(ts) }
func (c *LoanApplyController) Disburse(id core.Hash) error { return lpApply.Disburse(id) }
func (c *LoanApplyController) Get(id core.Hash) (core.LoanApplication, bool, error) {
	return lpApply.Get(id)
}
func (c *LoanApplyController) List(status core.ApplicationStatus) ([]core.LoanApplication, error) {
	return lpApply.List(status)
}

var loanApplyCmd = &cobra.Command{Use: "loanpool_apply", Short: "Loan application management", PersistentPreRunE: ensureLoanApply}

var laSubmitCmd = &cobra.Command{
	Use:  "submit <applicant> <amount> <termMonths> <purpose>",
	Args: cobra.MinimumNArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &LoanApplyController{}
		amt, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return err
		}
		term64, err := strconv.ParseUint(args[2], 10, 16)
		if err != nil {
			return err
		}
		id, err := ctrl.Submit(args[0], amt, uint16(term64), args[3])
		if err != nil {
			return err
		}
		fmt.Println("application", id.Hex())
		return nil
	},
}

var laVoteCmd = &cobra.Command{
	Use:  "vote <voter> <id> [--approve]",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &LoanApplyController{}
		b, err := hex.DecodeString(args[1])
		if err != nil {
			return err
		}
		var h core.Hash
		copy(h[:], b)
		approve, _ := cmd.Flags().GetBool("approve")
		return ctrl.Vote(args[0], h, approve)
	},
}

var laProcessCmd = &cobra.Command{
	Use:  "process",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &LoanApplyController{}
		ctrl.Process(time.Now().UTC())
		return nil
	},
}

var laDisburseCmd = &cobra.Command{
	Use:  "disburse <id>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &LoanApplyController{}
		b, err := hex.DecodeString(args[0])
		if err != nil {
			return err
		}
		var h core.Hash
		copy(h[:], b)
		return ctrl.Disburse(h)
	},
}

var laGetCmd = &cobra.Command{
	Use:  "get <id>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &LoanApplyController{}
		b, err := hex.DecodeString(args[0])
		if err != nil {
			return err
		}
		var h core.Hash
		copy(h[:], b)
		app, ok, err := ctrl.Get(h)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("not found")
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(app)
	},
}

var laListCmd = &cobra.Command{
	Use:  "list",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &LoanApplyController{}
		apps, err := ctrl.List(0)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(apps)
	},
}

func init() {
	laVoteCmd.Flags().Bool("approve", true, "approve or reject")
	loanApplyCmd.AddCommand(laSubmitCmd, laVoteCmd, laProcessCmd, laDisburseCmd, laGetCmd, laListCmd)
}

var LoanApplyCmd = loanApplyCmd
