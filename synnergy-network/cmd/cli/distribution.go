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
	distLedger *core.Ledger
	distCoin   *core.Coin
	dist       *core.Distributor
	distOnce   sync.Once
	distLogger = logrus.StandardLogger()
)

func distInit(cmd *cobra.Command, _ []string) error {
	var err error
	distOnce.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		distLedger, err = core.OpenLedger(path)
		if err != nil {
			return
		}
		distCoin, err = core.NewCoin(distLedger)
		if err != nil {
			return
		}
		dist = core.NewDistributor(distLedger, distCoin)
	})
	return err
}

func distParseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func distAirdrop(_ *cobra.Command, args []string) error {
	recips := make(map[core.Address]uint64)
	for _, pair := range args {
		parts := strings.Split(pair, ":")
		if len(parts) != 2 {
			return fmt.Errorf("bad recipient format")
		}
		addr, err := distParseAddr(parts[0])
		if err != nil {
			return err
		}
		amt, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return err
		}
		recips[addr] = amt
	}
	return dist.Airdrop(recips)
}

func distBatch(cmd *cobra.Command, args []string) error {
	fromStr, _ := cmd.Flags().GetString("from")
	from, err := distParseAddr(fromStr)
	if err != nil {
		return err
	}
	var items []core.TransferItem
	for _, pair := range args {
		parts := strings.Split(pair, ":")
		if len(parts) != 2 {
			return fmt.Errorf("bad recipient format")
		}
		addr, err := distParseAddr(parts[0])
		if err != nil {
			return err
		}
		amt, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return err
		}
		items = append(items, core.TransferItem{To: addr, Amount: amt})
	}
	return dist.BatchTransfer(from, items)
}

var distRootCmd = &cobra.Command{
	Use:               "distribution",
	Short:             "bulk token distribution",
	PersistentPreRunE: distInit,
}

var distAirdropCmd = &cobra.Command{
	Use:   "airdrop addr:amt [addr:amt...]",
	Short: "mint tokens to recipients",
	Args:  cobra.MinimumNArgs(1),
	RunE:  distAirdrop,
}

var distBatchCmd = &cobra.Command{
	Use:   "batch --from addr addr:amt [addr:amt...]",
	Short: "transfer from one address",
	Args:  cobra.MinimumNArgs(1),
	RunE:  distBatch,
}

func init() {
	distBatchCmd.Flags().String("from", "", "source address")
	_ = distBatchCmd.MarkFlagRequired("from")
	distRootCmd.AddCommand(distAirdropCmd, distBatchCmd)
}

var DistributionCmd = distRootCmd
