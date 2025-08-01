package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/wasmerio/wasmer-go/wasmer"

	"synnergy-network/core"
)

var (
	cmOnce    sync.Once
	cmManager *core.ContractManager
)

func cmInit(cmd *cobra.Command, _ []string) error {
	var err error
	cmOnce.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			path = "./ledger.db"
		}
		if err = core.InitLedger(path); err != nil {
			return
		}
		led := core.CurrentLedger()
		if core.GetContractRegistry() == nil {
			st, _ := core.NewInMemory()
			vm := core.NewHeavyVM(st, core.NewGasMeter(8_000_000), wasmer.NewEngine())
			core.InitContracts(led, vm)
		}
		cmManager = core.NewContractManager(led, core.GetContractRegistry())
	})
	return err
}

func cmHandleTransfer(cmd *cobra.Command, args []string) error {
	addrBytes, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}
	var addr core.Address
	copy(addr[:], addrBytes)
	newBytes, err := hex.DecodeString(args[1])
	if err != nil {
		return err
	}
	var newAddr core.Address
	copy(newAddr[:], newBytes)
	return cmManager.TransferOwnership(addr, newAddr)
}

func cmHandlePause(cmd *cobra.Command, args []string) error {
	b, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}
	var a core.Address
	copy(a[:], b)
	return cmManager.PauseContract(a)
}

func cmHandleResume(cmd *cobra.Command, args []string) error {
	b, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}
	var a core.Address
	copy(a[:], b)
	return cmManager.ResumeContract(a)
}

func cmHandleUpgrade(cmd *cobra.Command, args []string) error {
	b, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}
	var addr core.Address
	copy(addr[:], b)
	code, err := os.ReadFile(args[1])
	if err != nil {
		return err
	}
	gas, _ := cmd.Flags().GetUint64("gas")
	return cmManager.UpgradeContract(addr, code, gas)
}

func cmHandleInfo(cmd *cobra.Command, args []string) error {
	b, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}
	var a core.Address
	copy(a[:], b)
	info, err := cmManager.ContractInfo(a)
	if err != nil {
		return err
	}
	fmt.Println(string(info))
	return nil
}

var contractMgmtCmd = &cobra.Command{
	Use:               "contractops",
	Short:             "Manage deployed contracts",
	PersistentPreRunE: cmInit,
}

var cmTransferCmd = &cobra.Command{Use: "transfer <addr> <newOwner>", Args: cobra.ExactArgs(2), RunE: cmHandleTransfer}
var cmPauseCmd = &cobra.Command{Use: "pause <addr>", Args: cobra.ExactArgs(1), RunE: cmHandlePause}
var cmResumeCmd = &cobra.Command{Use: "resume <addr>", Args: cobra.ExactArgs(1), RunE: cmHandleResume}
var cmUpgradeCmd = &cobra.Command{Use: "upgrade <addr> <wasm>", Args: cobra.ExactArgs(2), RunE: cmHandleUpgrade}
var cmInfoCmd = &cobra.Command{Use: "info <addr>", Args: cobra.ExactArgs(1), RunE: cmHandleInfo}

func init() {
	cmUpgradeCmd.Flags().Uint64("gas", 200000, "gas limit")
	contractMgmtCmd.AddCommand(cmTransferCmd, cmPauseCmd, cmResumeCmd, cmUpgradeCmd, cmInfoCmd)
}

// ContractMgmtCmd exposes the root command.
var ContractMgmtCmd = contractMgmtCmd

func RegisterContractMgmt(root *cobra.Command) { root.AddCommand(ContractMgmtCmd) }
