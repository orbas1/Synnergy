package cli

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var empTok *core.EmploymentToken

func ensureEmpTok(cmd *cobra.Command, _ []string) error {
	if empTok != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return errors.New("ledger not initialised")
	}
	for _, t := range core.GetRegistryTokens() {
		if t.Meta().Standard == core.StdSYN3100 {
			if et, ok := any(t).(*core.EmploymentToken); ok {
				empTok = et
				break
			}
		}
	}
	if empTok == nil {
		return errors.New("employment token not found")
	}
	return nil
}

func empTokAdd(cmd *cobra.Command, args []string) error {
	employer, err := core.ParseAddress(args[1])
	if err != nil {
		return err
	}
	employee, err := core.ParseAddress(args[2])
	if err != nil {
		return err
	}
	salary, err := strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		return err
	}
	meta := core.EmploymentContractMeta{
		ContractID: args[0],
		Employer:   employer,
		Employee:   employee,
		Salary:     salary,
		Position:   cmd.Flag("position").Value.String(),
		Start:      time.Now().UTC(),
		Active:     true,
	}
	return empTok.CreateContract(meta)
}

func empTokPay(cmd *cobra.Command, args []string) error {
	return empTok.PaySalary(args[0])
}

func empTokShow(cmd *cobra.Command, args []string) error {
	meta, ok := empTok.GetContract(args[0])
	if !ok {
		return errors.New("not found")
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(meta)
}

var empTokenCmd = &cobra.Command{Use: "employmenttoken", Short: "Manage SYN3100 employment tokens", PersistentPreRunE: ensureEmpTok}

var empTokAddCmd = &cobra.Command{Use: "add <id> <employer> <employee> <salary>", Args: cobra.ExactArgs(4), RunE: empTokAdd}
var empTokPayCmd = &cobra.Command{Use: "pay <id>", Args: cobra.ExactArgs(1), RunE: empTokPay}
var empTokShowCmd = &cobra.Command{Use: "show <id>", Args: cobra.ExactArgs(1), RunE: empTokShow}

func init() {
	empTokAddCmd.Flags().String("position", "", "job position")
	empTokenCmd.AddCommand(empTokAddCmd, empTokPayCmd, empTokShowCmd)
}

var EmpTokenCmd = empTokenCmd
