package cli

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"synnergy-network/core"
)

func offwalletCreate(cmd *cobra.Command, _ []string) error {
	bits, _ := cmd.Flags().GetInt("bits")
	out, _ := cmd.Flags().GetString("out")
	if out == "" {
		return errors.New("--out required")
	}
	wallet, mnemonic, err := core.NewOffChainWallet(bits, nil)
	if err != nil {
		return err
	}
	seed := wallet.Seed()
	ks := map[string]string{"seed": fmt.Sprintf("%x", seed)}
	data, _ := json.MarshalIndent(ks, "", "  ")
	if err := os.WriteFile(out, data, 0o600); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "mnemonic: %s\n", mnemonic)
	return nil
}

func offwalletSign(cmd *cobra.Command, _ []string) error {
	walletFile, _ := cmd.Flags().GetString("wallet")
	inFile, _ := cmd.Flags().GetString("in")
	outFile, _ := cmd.Flags().GetString("out")
	acct, _ := cmd.Flags().GetUint32("account")
	idx, _ := cmd.Flags().GetUint32("index")
	gas, _ := cmd.Flags().GetUint64("gas")
	if walletFile == "" || inFile == "" {
		return errors.New("--wallet and --in required")
	}
	data, err := os.ReadFile(walletFile)
	if err != nil {
		return err
	}
	var ks map[string]string
	if err := json.Unmarshal(data, &ks); err != nil {
		return err
	}
	seed, err := hexToBytes(ks["seed"])
	if err != nil {
		return err
	}
	w, err := core.NewHDWalletFromSeed(seed, nil)
	if err != nil {
		return err
	}
	ow := &core.OffChainWallet{HDWallet: w}
	raw, err := os.ReadFile(inFile)
	if err != nil {
		return err
	}
	var tx core.Transaction
	if err := json.Unmarshal(raw, &tx); err != nil {
		return err
	}
	if err := ow.SignOffline(&tx, acct, idx, gas); err != nil {
		return err
	}
	out, _ := json.MarshalIndent(&tx, "", "  ")
	if outFile != "" {
		if err := os.WriteFile(outFile, out, 0o600); err != nil {
			return err
		}
	} else {
		cmd.OutOrStdout().Write(out)
		fmt.Fprintln(cmd.OutOrStdout())
	}
	return nil
}

func hexToBytes(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

var offWalletCmd = &cobra.Command{
	Use:   "offwallet",
	Short: "Offline wallet utilities",
}

var offCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create offline wallet",
	RunE:  offwalletCreate,
}

var offSignCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign a transaction offline",
	RunE:  offwalletSign,
}

func init() {
	offCreateCmd.Flags().Int("bits", 128, "entropy bits")
	offCreateCmd.Flags().String("out", "", "wallet output file")

	offSignCmd.Flags().String("wallet", "", "wallet file")
	offSignCmd.Flags().String("in", "", "unsigned tx")
	offSignCmd.Flags().String("out", "", "signed tx output")
	offSignCmd.Flags().Uint32("account", 0, "account")
	offSignCmd.Flags().Uint32("index", 0, "index")
	offSignCmd.Flags().Uint64("gas", 0, "gas price")

	offWalletCmd.AddCommand(offCreateCmd, offSignCmd)
}

var OffWalletCmd = offWalletCmd

func RegisterOffWallet(root *cobra.Command) { root.AddCommand(offWalletCmd) }
