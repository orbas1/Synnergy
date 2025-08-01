package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"synnergy-network/core"
)

var (
	spOnce   sync.Once
	spLogger = logrus.StandardLogger()
	spMgr    *core.StakePenaltyManager
)

func spInitMiddleware(cmd *cobra.Command, _ []string) error {
	var err error
	spOnce.Do(func() {
		_ = godotenv.Load()
		lvl := os.Getenv("LOG_LEVEL")
		if lvl == "" {
			lvl = "info"
		}
		l, e := logrus.ParseLevel(lvl)
		if e != nil {
			err = e
			return
		}
		spLogger.SetLevel(l)
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		led, e := core.OpenLedger(path)
		if e != nil {
			err = e
			return
		}
		spMgr = core.NewStakePenaltyManager(spLogger, led)
	})
	return err
}

func spParseAddr(s string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func spHandleAdjust(cmd *cobra.Command, args []string) error {
	addr, err := spParseAddr(args[0])
	if err != nil {
		return err
	}
	delta, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return err
	}
	return spMgr.AdjustStake(addr, delta)
}

func spHandlePenalty(cmd *cobra.Command, args []string) error {
	addr, err := spParseAddr(args[0])
	if err != nil {
		return err
	}
	pts, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil {
		return err
	}
	reason := ""
	if len(args) > 2 {
		reason = args[2]
	}
	return spMgr.Penalize(addr, uint32(pts), reason)
}

func spHandleInfo(cmd *cobra.Command, args []string) error {
	addr, err := spParseAddr(args[0])
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "stake=%d penalty=%d\n", spMgr.StakeOf(addr), spMgr.PenaltyOf(addr))
	return nil
}

var stakeCmd = &cobra.Command{
	Use:               "stake",
	Short:             "Manage validator stakes and penalties",
	PersistentPreRunE: spInitMiddleware,
}

var stakeAdjustCmd = &cobra.Command{Use: "adjust <addr> <delta>", Args: cobra.ExactArgs(2), RunE: spHandleAdjust}
var stakePenaltyCmd = &cobra.Command{Use: "penalize <addr> <points> [reason]", Args: cobra.RangeArgs(2, 3), RunE: spHandlePenalty}
var stakeInfoCmd = &cobra.Command{Use: "info <addr>", Args: cobra.ExactArgs(1), RunE: spHandleInfo}

func init() {
	stakeCmd.AddCommand(stakeAdjustCmd, stakePenaltyCmd, stakeInfoCmd)
}

var StakeCmd = stakeCmd
