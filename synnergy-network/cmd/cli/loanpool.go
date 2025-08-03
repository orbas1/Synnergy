package cli

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	loanPool *core.LoanPool
)

type lpElector struct{}

func (lpElector) RandomElectorate(n int) ([]core.Address, error) {
	return nil, errors.New("not implemented")
}
func (lpElector) IsAuthority(a core.Address) bool { return true }

func ensureLoanPool(cmd *cobra.Command, _ []string) error {
	if loanPool != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return errors.New("ledger not initialised")
	}
	std := log.New(logrus.StandardLogger().Out, "", log.LstdFlags)
	loanPool = core.NewLoanPool(std, led, lpElector{}, &core.LoanPoolConfig{})
	return nil
}

// Controller thin wrapper

type LoanPoolController struct{}

func (c *LoanPoolController) Submit(creator, recip string, t core.ProposalType, amt uint64, desc string) (core.Hash, error) {
	ca, err := core.StringToAddress(creator)
	if err != nil {
		return core.Hash{}, err
	}
	ra, err := core.StringToAddress(recip)
	if err != nil {
		return core.Hash{}, err
	}
	return loanPool.Submit(ca, ra, t, amt, desc)
}

func (c *LoanPoolController) Vote(voter string, id core.Hash, approve bool) error {
	v, err := core.StringToAddress(voter)
	if err != nil {
		return err
	}
	return loanPool.Vote(v, id, approve)
}

func (c *LoanPoolController) Disburse(id core.Hash) error { return loanPool.Disburse(id) }
func (c *LoanPoolController) Tick(ts time.Time)           { loanPool.Tick(ts) }
func (c *LoanPoolController) Get(id core.Hash) (core.Proposal, bool, error) {
	return loanPool.GetProposal(id)
}
func (c *LoanPoolController) List(status core.ProposalStatus) ([]core.Proposal, error) {
	return loanPool.ListProposals(status)
}
func (c *LoanPoolController) Cancel(creator string, id core.Hash) error {
	return loanPool.CancelProposal(id, core.Address(creator))
}
func (c *LoanPoolController) Extend(creator string, id core.Hash, hrs int) error {
	return loanPool.ExtendProposal(id, time.Duration(hrs)*time.Hour, core.Address(creator))
}

// CLI commands

var loanCmd = &cobra.Command{Use: "loanpool", Short: "Loan pool operations", PersistentPreRunE: ensureLoanPool}

var lpSubmitCmd = &cobra.Command{
	Use: "submit <creator> <recipient> <type> <amount> <desc>", Args: cobra.MinimumNArgs(5),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &LoanPoolController{}
		typ, err := strconv.Atoi(args[2])
		if err != nil {
			return err
		}
		amt, err := strconv.ParseUint(args[3], 10, 64)
		if err != nil {
			return err
		}
		id, err := ctrl.Submit(args[0], args[1], core.ProposalType(typ), amt, args[4])
		if err != nil {
			return err
		}
		logrus.WithField("proposal_id", id.Hex()).Info("proposal submitted")
		return nil
	},
}

var lpVoteCmd = &cobra.Command{
	Use: "vote <voter> <id> [--approve]", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &LoanPoolController{}
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

var lpDisburseCmd = &cobra.Command{Use: "disburse <id>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	ctrl := &LoanPoolController{}
	b, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}
	var h core.Hash
	copy(h[:], b)
	return ctrl.Disburse(h)
}}

var lpTickCmd = &cobra.Command{Use: "tick", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
	ctrl := &LoanPoolController{}
	ctrl.Tick(time.Now().UTC())
	return nil
}}

var lpGetCmd = &cobra.Command{Use: "get <id>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	ctrl := &LoanPoolController{}
	b, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}
	var h core.Hash
	copy(h[:], b)
	p, ok, err := ctrl.Get(h)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("not found")
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}}

var lpListCmd = &cobra.Command{Use: "list", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
	ctrl := &LoanPoolController{}
	props, err := ctrl.List(0)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(props)
}}

var lpCancelCmd = &cobra.Command{Use: "cancel <creator> <id>", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
	ctrl := &LoanPoolController{}
	b, err := hex.DecodeString(args[1])
	if err != nil {
		return err
	}
	var h core.Hash
	copy(h[:], b)
	return ctrl.Cancel(args[0], h)
}}

var lpExtendCmd = &cobra.Command{Use: "extend <creator> <id> <hrs>", Args: cobra.ExactArgs(3), RunE: func(cmd *cobra.Command, args []string) error {
	ctrl := &LoanPoolController{}
	b, err := hex.DecodeString(args[1])
	if err != nil {
		return err
	}
	var h core.Hash
	copy(h[:], b)
	hrs, err := strconv.Atoi(args[2])
	if err != nil {
		return err
	}
	return ctrl.Extend(args[0], h, hrs)
}}

func init() {
	lpVoteCmd.Flags().Bool("approve", true, "approve or reject")
	loanCmd.AddCommand(
		lpSubmitCmd,
		lpVoteCmd,
		lpDisburseCmd,
		lpTickCmd,
		lpGetCmd,
		lpListCmd,
		lpCancelCmd,
		lpExtendCmd,
	)
}

var LoanCmd = loanCmd
