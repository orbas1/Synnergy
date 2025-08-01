package cli

import (
	"encoding/hex"
	"fmt"
	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var custNode *core.CustodialNode

func custodialInit(cmd *cobra.Command, _ []string) error {
	if custNode != nil {
		return nil
	}
	cfg := core.CustodialConfig{
		Network: core.Config{ListenAddr: "/ip4/0.0.0.0/tcp/4100"},
		Ledger:  core.LedgerConfig{},
	}
	n, err := core.NewCustodialNode(cfg)
	if err != nil {
		return err
	}
	custNode = n
	return nil
}

func custodialStart(cmd *cobra.Command, _ []string) error {
	if custNode == nil {
		return fmt.Errorf("not initialised")
	}
	custNode.Start()
	fmt.Fprintln(cmd.OutOrStdout(), "custodial node started")
	return nil
}

func custodialStop(cmd *cobra.Command, _ []string) error {
	if custNode == nil {
		return nil
	}
	_ = custNode.Stop()
	custNode = nil
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func custodialDeposit(cmd *cobra.Command, args []string) error {
	if custNode == nil {
		return fmt.Errorf("not running")
	}
	addr := hexToAddr(args[0])
	tid, err := parseUint(args[1])
	if err != nil {
		return err
	}
	token := core.TokenID(tid)
	amt, err := parseUint(args[2])
	if err != nil {
		return err
	}
	return custNode.Deposit(addr, token, amt)
}

func custodialWithdraw(cmd *cobra.Command, args []string) error {
	if custNode == nil {
		return fmt.Errorf("not running")
	}
	addr := hexToAddr(args[0])
	tid, err := parseUint(args[1])
	if err != nil {
		return err
	}
	token := core.TokenID(tid)
	amt, err := parseUint(args[2])
	if err != nil {
		return err
	}
	return custNode.Withdraw(addr, token, amt)
}

func custodialBalance(cmd *cobra.Command, args []string) error {
	if custNode == nil {
		return fmt.Errorf("not running")
	}
	addr := hexToAddr(args[0])
	tid, err := parseUint(args[1])
	if err != nil {
		return err
	}
	token := core.TokenID(tid)
	bal := custNode.BalanceOf(addr, token)
	fmt.Fprintf(cmd.OutOrStdout(), "%d\n", bal)
	return nil
}

func hexToAddr(s string) core.Address {
	var a core.Address
	b, _ := hex.DecodeString(s)
	copy(a[:], b)
	return a
}

func parseUint(s string) (uint64, error) {
	var n uint64
	_, err := fmt.Sscan(s, &n)
	return n, err
}

var custCmd = &cobra.Command{Use: "custodial", Short: "Custodial node", PersistentPreRunE: custodialInit}
var custStartCmd = &cobra.Command{Use: "start", Short: "Start", RunE: custodialStart}
var custStopCmd = &cobra.Command{Use: "stop", Short: "Stop", RunE: custodialStop}
var custDepositCmd = &cobra.Command{Use: "deposit <addr> <token> <amt>", Short: "Deposit asset", Args: cobra.ExactArgs(3), RunE: custodialDeposit}
var custWithdrawCmd = &cobra.Command{Use: "withdraw <addr> <token> <amt>", Short: "Withdraw asset", Args: cobra.ExactArgs(3), RunE: custodialWithdraw}
var custBalanceCmd = &cobra.Command{Use: "balance <addr> <token>", Short: "Balance", Args: cobra.ExactArgs(2), RunE: custodialBalance}

func init() {
	custCmd.AddCommand(custStartCmd, custStopCmd, custDepositCmd, custWithdrawCmd, custBalanceCmd)
}

var CustodialCmd = custCmd
