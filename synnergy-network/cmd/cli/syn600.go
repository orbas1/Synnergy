package cli

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	syn600Once   sync.Once
	syn600Ledger *core.Ledger
	syn600Token  *core.SYN600Token
)

const syn600ID = core.TokenID(0x53000000 | uint32(core.StdSYN600)<<8)

func syn600Init(cmd *cobra.Command, _ []string) error {
	var err error
	syn600Once.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		syn600Ledger, err = core.OpenLedger(path)
		if err != nil {
			return
		}
		tok, ok := core.GetToken(syn600ID)
		if !ok {
			err = fmt.Errorf("SYN600 token not registered")
			return
		}
		syn600Token = core.NewSYN600Token(tok.(*core.BaseToken), syn600Ledger)
	})
	return err
}

func syn600Stake(cmd *cobra.Command, args []string) error {
	addr, err := core.StringToAddress(args[0])
	if err != nil {
		return err
	}
	amt, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	days, _ := cmd.Flags().GetInt("days")
	if days <= 0 {
		days = 1
	}
	return syn600Token.Stake(addr, amt, time.Duration(days)*24*time.Hour)
}

func syn600Unstake(cmd *cobra.Command, args []string) error {
	addr, err := core.StringToAddress(args[0])
	if err != nil {
		return err
	}
	return syn600Token.Unstake(addr)
}

func syn600Reward(cmd *cobra.Command, args []string) error {
	addr, err := core.StringToAddress(args[0])
	if err != nil {
		return err
	}
	amt, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	return syn600Token.Mint(addr, amt)
}

func syn600Engage(cmd *cobra.Command, args []string) error {
	addr, err := core.StringToAddress(args[0])
	if err != nil {
		return err
	}
	pts, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	return syn600Token.AddEngagement(addr, pts)
}

func syn600Engagement(cmd *cobra.Command, args []string) error {
	addr, err := core.StringToAddress(args[0])
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%d\n", syn600Token.EngagementOf(addr))
	return nil
}

var syn600Cmd = &cobra.Command{
	Use:               "syn600",
	Short:             "SYN600 reward token operations",
	PersistentPreRunE: syn600Init,
}

var syn600StakeCmd = &cobra.Command{Use: "stake <addr> <amt>", Short: "stake tokens", Args: cobra.ExactArgs(2), RunE: syn600Stake}
var syn600UnstakeCmd = &cobra.Command{Use: "unstake <addr>", Short: "unstake tokens", Args: cobra.ExactArgs(1), RunE: syn600Unstake}
var syn600RewardCmd = &cobra.Command{Use: "reward <addr> <amt>", Short: "mint reward", Args: cobra.ExactArgs(2), RunE: syn600Reward}
var syn600EngageCmd = &cobra.Command{Use: "engage <addr> <pts>", Short: "record engagement", Args: cobra.ExactArgs(2), RunE: syn600Engage}
var syn600EngagementCmd = &cobra.Command{Use: "engagement <addr>", Short: "show engagement", Args: cobra.ExactArgs(1), RunE: syn600Engagement}

func init() {
	syn600StakeCmd.Flags().Int("days", 1, "lock duration in days")
	syn600Cmd.AddCommand(syn600StakeCmd, syn600UnstakeCmd, syn600RewardCmd, syn600EngageCmd, syn600EngagementCmd)
}

var SYN600Cmd = syn600Cmd

func RegisterSYN600(root *cobra.Command) { root.AddCommand(SYN600Cmd) }
