package cli

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	stakeOnce   sync.Once
	stakeLedger *core.Ledger
	stakeMgr    *core.DAOStaking
)

func stakeInit(cmd *cobra.Command, _ []string) error {
	var err error
	stakeOnce.Do(func() {
		_ = godotenv.Load()
		wal := envOr("LEDGER_WAL", "./ledger.wal")
		snap := envOr("LEDGER_SNAPSHOT", "./ledger.snap")
		interval := envOrInt("LEDGER_SNAPSHOT_INTERVAL", 100)
		stakeLedger, err = core.NewLedger(core.LedgerConfig{
			WALPath:          wal,
			SnapshotPath:     snap,
			SnapshotInterval: interval,
		})
		if err != nil {
			return
		}
		core.InitDAOStaking(logrus.StandardLogger(), stakeLedger)
		stakeMgr = core.StakingManager()
	})
	return err
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envOrInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// -----------------------------------------------------------------------------
// Commands
// -----------------------------------------------------------------------------

var stakeCmd = &cobra.Command{Use: "staking", Short: "DAO staking", PersistentPreRunE: stakeInit}

var stakeDoCmd = &cobra.Command{
	Use:   "stake <addr> <amount>",
	Short: "Lock tokens for governance",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := core.StringToAddress(args[0])
		if err != nil {
			return err
		}
		amt, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return err
		}
		return stakeMgr.Stake(addr, amt)
	},
}

var stakeUnCmd = &cobra.Command{
	Use:   "unstake <addr> <amount>",
	Short: "Release staked tokens",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := core.StringToAddress(args[0])
		if err != nil {
			return err
		}
		amt, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return err
		}
		return stakeMgr.Unstake(addr, amt)
	},
}

var stakeBalCmd = &cobra.Command{
	Use:   "balance <addr>",
	Short: "Show staked balance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := core.StringToAddress(args[0])
		if err != nil {
			return err
		}
		fmt.Println(stakeMgr.StakedOf(addr))
		return nil
	},
}

var stakeTotalCmd = &cobra.Command{
	Use:   "total",
	Short: "Total staked tokens",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(stakeMgr.TotalStaked())
		return nil
	},
}

func init() { stakeCmd.AddCommand(stakeDoCmd, stakeUnCmd, stakeBalCmd, stakeTotalCmd) }

// StakingCmd exported for CLI index
var StakingCmd = stakeCmd

func RegisterStaking(root *cobra.Command) { root.AddCommand(StakingCmd) }
