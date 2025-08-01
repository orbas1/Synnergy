package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var empCtrl *EmploymentController

func ensureEmployment(cmd *cobra.Command, _ []string) error {
	if empCtrl != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return errors.New("ledger not initialised")
	}
	core.InitEmployment(led)
	empCtrl = &EmploymentController{reg: core.Employment()}
	return nil
}

type EmploymentController struct{ reg *core.EmploymentRegistry }

func (c *EmploymentController) Create(employer, employee string, salary uint64, hrs int) (string, error) {
	emp, err := core.ParseAddress(employer)
	if err != nil {
		return "", err
	}
	worker, err := core.ParseAddress(employee)
	if err != nil {
		return "", err
	}
	start := time.Now().UTC()
	end := start.Add(time.Duration(hrs) * time.Hour)
	return c.reg.CreateJob(emp, worker, salary, start, end)
}

func (c *EmploymentController) Sign(id, addr string) error {
	a, err := core.ParseAddress(addr)
	if err != nil {
		return err
	}
	return c.reg.SignJob(id, a)
}

func (c *EmploymentController) Hours(id string, h uint32) error {
	return c.reg.RecordWork(id, h)
}

func (c *EmploymentController) Pay(id string) error { return c.reg.PaySalary(id) }

func (c *EmploymentController) Show(id string) (core.EmploymentContract, error) {
	ec, ok, err := c.reg.GetJob(id)
	if err != nil {
		return core.EmploymentContract{}, err
	}
	if !ok {
		return core.EmploymentContract{}, errors.New("not found")
	}
	return ec, nil
}

var employmentCmd = &cobra.Command{Use: "employment", Short: "Employment contracts", PersistentPreRunE: ensureEmployment}

var empCreateCmd = &cobra.Command{Use: "create <employer> <employee> <salary> <hours>", Args: cobra.ExactArgs(4), RunE: func(cmd *cobra.Command, args []string) error {
	sal, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return err
	}
	hrs, err := strconv.Atoi(args[3])
	if err != nil {
		return err
	}
	id, err := empCtrl.Create(args[0], args[1], sal, hrs)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), id)
	return nil
}}

var empSignCmd = &cobra.Command{Use: "sign <id> <addr>", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
	return empCtrl.Sign(args[0], args[1])
}}

var empHoursCmd = &cobra.Command{Use: "hours <id> <n>", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
	n, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil {
		return err
	}
	return empCtrl.Hours(args[0], uint32(n))
}}

var empPayCmd = &cobra.Command{Use: "pay <id>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	return empCtrl.Pay(args[0])
}}

var empShowCmd = &cobra.Command{Use: "show <id>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	ec, err := empCtrl.Show(args[0])
	if err != nil {
		return err
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(ec)
}}

func init() {
	employmentCmd.AddCommand(empCreateCmd, empSignCmd, empHoursCmd, empPayCmd, empShowCmd)
}

var EmploymentCmd = employmentCmd
