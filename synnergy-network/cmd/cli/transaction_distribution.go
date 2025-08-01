package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	core "synnergy-network/core"
)

var txdCmd = &cobra.Command{
	Use:   "txdist",
	Short: "Distribute transaction fees",
}

var txdDistributeCmd = &cobra.Command{
	Use:   "distribute",
	Short: "simulate fee distribution",
	RunE: func(cmd *cobra.Command, args []string) error {
		fromStr := viper.GetString("from")
		minerStr := viper.GetString("miner")
		feeStr := viper.GetString("fee")
		if fromStr == "" || minerStr == "" || feeStr == "" {
			return fmt.Errorf("from, miner and fee required")
		}
		from, err := core.StringToAddress(fromStr)
		if err != nil {
			return err
		}
		minerBytes, err := hexToBytes(minerStr)
		if err != nil {
			return err
		}
		fee, err := strconv.ParseUint(feeStr, 10, 64)
		if err != nil {
			return err
		}
		dist := core.CurrentTxDistributor()
		if dist == nil {
			return fmt.Errorf("distributor not initialised")
		}
		if err := dist.DistributeFees(from, minerBytes, fee); err != nil {
			return err
		}
		fmt.Println("distribution complete")
		return nil
	},
}

func init() {
	txdDistributeCmd.Flags().String("from", "", "sender address hex")
	txdDistributeCmd.Flags().String("miner", "", "miner public key hex")
	txdDistributeCmd.Flags().String("fee", "", "fee amount")
	_ = viper.BindPFlags(txdDistributeCmd.Flags())

	txdCmd.AddCommand(txdDistributeCmd)
}

// TxDistributionCmd exposes the root command.
func TxDistributionCmd() *cobra.Command { return txdCmd }

func hexToBytes(str string) ([]byte, error) {
	str = strings.TrimPrefix(str, "0x")
	return hex.DecodeString(str)
}
