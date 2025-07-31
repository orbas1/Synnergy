package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	core "synnergy-network/core"
)

// ensureAMMInitialised reused from amm.go
func lpEnsureInit(cmd *cobra.Command, _ []string) error {
	if core.Manager() != nil {
		return nil
	}
	fixture := viper.GetString("AMM_POOLS_FIXTURE")
	if fixture == "" {
		return fmt.Errorf("AMM manager not initialised – set AMM_POOLS_FIXTURE")
	}
	return core.InitPoolsFromFile(fixture)
}

type lpController struct{}

func (lpController) Create(a, b core.TokenID, fee uint16) (core.PoolID, error) {
	return core.Manager().CreatePool(a, b, fee)
}

func (lpController) Add(pid core.PoolID, provider core.Address, aAmt, bAmt uint64) (uint64, error) {
	return core.Manager().AddLiquidity(pid, provider, aAmt, bAmt)
}

func (lpController) Swap(pid core.PoolID, trader core.Address, inTok core.TokenID, inAmt, minOut uint64) (uint64, error) {
	return core.Manager().Swap(pid, trader, inTok, inAmt, minOut)
}

func (lpController) Remove(pid core.PoolID, provider core.Address, lp uint64) (uint64, uint64, error) {
	return core.Manager().RemoveLiquidity(pid, provider, lp)
}

func (lpController) Pool(pid core.PoolID) (core.Pool, error) { return core.Manager().Pool(pid) }
func (lpController) Pools() []core.Pool                      { return core.Manager().Pools() }

func mustAddr(hexStr string) core.Address {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(hexStr, "0x"))
	if err != nil || len(b) != len(a) {
		return a
	}
	copy(a[:], b)
	return a
}

var poolsCmd = &cobra.Command{Use: "pools", Short: "Manage liquidity pools", PersistentPreRunE: lpEnsureInit}

var poolCreateCmd = &cobra.Command{
	Use:   "create <tokenA> <tokenB> [feeBps]",
	Short: "Create a new pool",
	Args:  cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctl := lpController{}
		tA, tB := core.TokenID(args[0]), core.TokenID(args[1])
		fee := uint16(0)
		if len(args) == 3 {
			f, err := strconv.ParseUint(args[2], 10, 16)
			if err != nil {
				return err
			}
			fee = uint16(f)
		}
		pid, err := ctl.Create(tA, tB, fee)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%d\n", pid)
		return nil
	},
}

var poolAddCmd = &cobra.Command{
	Use:   "add <poolID> <provider> <amtA> <amtB>",
	Short: "Add liquidity to a pool",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctl := lpController{}
		pid := core.PoolID(args[0])
		provider := mustAddr(args[1])
		aAmt, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return err
		}
		bAmt, err := strconv.ParseUint(args[3], 10, 64)
		if err != nil {
			return err
		}
		minted, err := ctl.Add(pid, provider, aAmt, bAmt)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%d\n", minted)
		return nil
	},
}

var poolSwapCmd = &cobra.Command{
	Use:   "swap <poolID> <trader> <tokenIn> <amtIn> <minOut>",
	Short: "Swap tokens within a pool",
	Args:  cobra.ExactArgs(5),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctl := lpController{}
		pid := core.PoolID(args[0])
		trader := mustAddr(args[1])
		inTok := core.TokenID(args[2])
		inAmt, err := strconv.ParseUint(args[3], 10, 64)
		if err != nil {
			return err
		}
		minOut, err := strconv.ParseUint(args[4], 10, 64)
		if err != nil {
			return err
		}
		out, err := ctl.Swap(pid, trader, inTok, inAmt, minOut)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%d\n", out)
		return nil
	},
}

var poolRemoveCmd = &cobra.Command{
	Use:   "remove <poolID> <provider> <lpTokens>",
	Short: "Remove liquidity from a pool",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctl := lpController{}
		pid := core.PoolID(args[0])
		provider := mustAddr(args[1])
		lpAmt, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return err
		}
		a, b, err := ctl.Remove(pid, provider, lpAmt)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%d %d\n", a, b)
		return nil
	},
}

var poolInfoCmd = &cobra.Command{
	Use:   "info <poolID>",
	Short: "Show pool state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctl := lpController{}
		pid := core.PoolID(args[0])
		p, err := ctl.Pool(pid)
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(p, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(enc))
		return nil
	},
}

var poolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all pools",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctl := lpController{}
		pools := ctl.Pools()
		enc, _ := json.MarshalIndent(pools, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(enc))
		return nil
	},
}

func init() {
	poolsCmd.AddCommand(poolCreateCmd, poolAddCmd, poolSwapCmd, poolRemoveCmd, poolInfoCmd, poolListCmd)
}

var PoolsCmd = poolsCmd
