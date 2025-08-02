package cli

// ──────────────────────────────────────────────────────────────────────────────
// Synthron Smart‑Contracts CLI
//
// Root command:          `contracts`
// Sub‑routes (micro‑CLIs):
//   compile     – compile .wat/.wasm → deterministic wasm blob
//   deploy      – deploy contract byte‑code + ricardian JSON to ledger
//   invoke      – call method with arbitrary args (hex) + gas limit
//   list        – list deployed contract addresses & code hash
//   info        – show ricardian manifest for address
//
// Layout rules honored:
//   • Command objects declared first; export consolidated at bottom.
//   • PersistentPreRunE wires middleware once (ledger, VM, registry).
//   • Controllers implement business logic with robust error handling.
//
// Env variables (add to .env):
//   LEDGER_PATH     – path to ledger db (required)
//   WASM_OUT_DIR    – directory for compiled wasm artifacts (default ./wasm)
//   LOG_LEVEL       – trace|debug|info|warn|error (default info)
//
// Usage examples after hooking into root CLI:
//   ~contracts ~compile ./hello.wat                   # → prints hash & path
//   ~contracts ~deploy --wasm ./hello.wasm --ric ./manifest.json --gas 5_000_000
//   ~contracts ~list
//   ~contracts ~invoke 0xabc... --method greet --args 48656c6c6f --gas 200_000
//
// ──────────────────────────────────────────────────────────────────────────────

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/wasmerio/wasmer-go/wasmer"

	"synnergy-network/core"
)

// ──────────────────────────────────────────────────────────────────────────────
// Globals + lazy‑Init middleware
// ──────────────────────────────────────────────────────────────────────────────

var (
	contractsLedger *core.Ledger
	contractsLogger = logrus.StandardLogger()
	contractsOnce   sync.Once
	vmSvc           core.VM
)

func initContractsMiddleware(cmd *cobra.Command, _ []string) error {
	var err error
	contractsOnce.Do(func() {
		_ = godotenv.Load()

		lvlStr := os.Getenv("LOG_LEVEL")
		if lvlStr == "" {
			lvlStr = "info"
		}
		lvl, e := logrus.ParseLevel(lvlStr)
		if e != nil {
			err = fmt.Errorf("invalid LOG_LEVEL: %w", e)
			return
		}
		contractsLogger.SetLevel(lvl)

		lp := os.Getenv("LEDGER_PATH")
		if lp == "" {
			err = fmt.Errorf("LEDGER_PATH env not set")
			return
		}
		if e := core.InitLedger(lp); e != nil {
			err = fmt.Errorf("open ledger: %w", e)
			return
		}
		contractsLedger = core.CurrentLedger()
		state, _ := core.NewInMemory()
		vmSvc = core.NewHeavyVM(state, core.NewGasMeter(8_000_000), wasmer.NewEngine())
		core.InitContracts(contractsLedger, vmSvc)
	})
	return err
}

// ──────────────────────────────────────────────────────────────────────────────
// Helper utilities
// ──────────────────────────────────────────────────────────────────────────────

func mustParseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("invalid address %s", h)
	}
	copy(a[:], b)
	return a, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Controllers
// ──────────────────────────────────────────────────────────────────────────────

type compileFlags struct{ src string }

func handleCompile(cmd *cobra.Command, _ []string) error {
	cf := cmd.Context().Value("cflags").(compileFlags)
	outDir := os.Getenv("WASM_OUT_DIR")
	if outDir == "" {
		outDir = "./wasm"
	}
	_ = os.MkdirAll(outDir, 0o755)

	wasm, hash, err := core.CompileWASM(cf.src, outDir)
	if err != nil {
		return err
	}

	outPath := filepath.Join(outDir, fmt.Sprintf("%x.wasm", hash[:8]))
	if err := os.WriteFile(outPath, wasm, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "compiled → %s\nhash: %x\n", outPath, hash[:])
	return nil
}

type deployFlags struct {
	wasm string
	ric  string
	gas  uint64
}

func handleDeploy(cmd *cobra.Command, _ []string) error {
	df := cmd.Context().Value("dflags").(deployFlags)

	code, err := os.ReadFile(df.wasm)
	if err != nil {
		return err
	}
	var ricData []byte
	if df.ric != "" {
		ricData, err = os.ReadFile(df.ric)
		if err != nil {
			return err
		}
	}

	// derive address & register
	caller := core.Address{} // system account 0x0…; could be flag in future
	addr := core.DeriveContractAddress(caller, code)
	cr := core.GetContractRegistry()
	if err := cr.Deploy(addr, code, ricData, df.gas); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "deployed at 0x%x\n", addr[:])
	return nil
}

type invokeFlags struct {
	method string
	args   string
	gas    uint64
}

type debugFlags struct {
	method string
	args   string
	gas    uint64
}

func handleInvoke(cmd *cobra.Command, args []string) error {
	addrStr := args[0]
	inv := cmd.Context().Value("iflags").(invokeFlags)

	addr, err := mustParseAddr(addrStr)
	if err != nil {
		return err
	}

	argBytes, err := hex.DecodeString(strings.TrimPrefix(inv.args, "0x"))
	if err != nil && inv.args != "" {
		return fmt.Errorf("args must be hex bytes")
	}

	caller := core.Address{} // could add flag later
	out, err := core.GetContractRegistry().Invoke(caller, addr, inv.method, argBytes, inv.gas)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%x\n", out)
	return nil
}

func handleDebug(cmd *cobra.Command, args []string) error {
	addrStr := args[0]
	df := cmd.Context().Value("dbgflags").(debugFlags)

	addr, err := mustParseAddr(addrStr)
	if err != nil {
		return err
	}

	argBytes, err := hex.DecodeString(strings.TrimPrefix(df.args, "0x"))
	if err != nil && df.args != "" {
		return fmt.Errorf("args must be hex bytes")
	}

	caller := core.Address{}
	rec, err := core.GetContractRegistry().InvokeWithReceipt(caller, addr, df.method, argBytes, df.gas)
	if err != nil {
		return err
	}
	b, _ := json.MarshalIndent(rec, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(b))
	return nil
}

func handleList(cmd *cobra.Command, _ []string) error {
	for addr, sc := range core.GetContractRegistry().All() {
		fmt.Fprintf(cmd.OutOrStdout(), "0x%x\t%x\tgas %d\n", addr[:], sc.CodeHash[:8], sc.GasLimit)
	}
	return nil
}

func handleInfo(cmd *cobra.Command, args []string) error {
	addr, err := mustParseAddr(args[0])
	if err != nil {
		return err
	}
	ric, err := core.GetContractRegistry().Ricardian(addr)
	if err != nil {
		return err
	}
	var pretty map[string]any
	if err := json.Unmarshal(ric, &pretty); err == nil {
		b, _ := json.MarshalIndent(pretty, "", "  ")
		cmd.OutOrStdout().Write(b)
		fmt.Fprintln(cmd.OutOrStdout())
	} else {
		cmd.OutOrStdout().Write(ric)
		fmt.Fprintln(cmd.OutOrStdout())
	}
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Cobra command tree
// ──────────────────────────────────────────────────────────────────────────────

var contractsCmd = &cobra.Command{
	Use:               "contracts",
	Short:             "Compile, deploy & invoke WASM smart‑contracts",
	PersistentPreRunE: initContractsMiddleware,
}

var compileCmd = &cobra.Command{
	Use:   "compile <src.wat|src.wasm>",
	Short: "Compile WAT/wasm to deterministic wasm blob",
	Args:  cobra.ExactArgs(1),
	RunE:  handleCompile,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		cf := compileFlags{src: args[0]}
		cmd.SetContext(context.WithValue(cmd.Context(), "cflags", cf))
		return nil
	},
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy compiled wasm with optional ricardian JSON",
	Args:  cobra.NoArgs,
	RunE:  handleDeploy,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		df := deployFlags{}
		df.wasm, _ = cmd.Flags().GetString("wasm")
		df.ric, _ = cmd.Flags().GetString("ric")
		gasStr, _ := cmd.Flags().GetString("gas")
		if df.wasm == "" {
			return fmt.Errorf("--wasm required")
		}
		if gasStr == "" {
			df.gas = 3_000_000
		} else {
			g, err := strconv.ParseUint(gasStr, 10, 64)
			if err != nil {
				return err
			}
			df.gas = g
		}
		cmd.SetContext(context.WithValue(cmd.Context(), "dflags", df))
		return nil
	},
}

var invokeCmd = &cobra.Command{
	Use:   "invoke <address>",
	Short: "Invoke a contract method",
	Args:  cobra.ExactArgs(1),
	RunE:  handleInvoke,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		iv := invokeFlags{}
		iv.method, _ = cmd.Flags().GetString("method")
		iv.args, _ = cmd.Flags().GetString("args")
		iv.gas, _ = cmd.Flags().GetUint64("gas")
		if iv.method == "" {
			return fmt.Errorf("--method required")
		}
		cmd.SetContext(context.WithValue(cmd.Context(), "iflags", iv))
		return nil
	},
}

var debugCmd = &cobra.Command{
	Use:   "debug <address>",
	Short: "Invoke contract and print full receipt",
	Args:  cobra.ExactArgs(1),
	RunE:  handleDebug,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		df := debugFlags{}
		df.method, _ = cmd.Flags().GetString("method")
		df.args, _ = cmd.Flags().GetString("args")
		df.gas, _ = cmd.Flags().GetUint64("gas")
		if df.method == "" {
			return fmt.Errorf("--method required")
		}
		cmd.SetContext(context.WithValue(cmd.Context(), "dbgflags", df))
		return nil
	},
}

var contractsListCmd = &cobra.Command{Use: "list", Short: "List deployed contracts", Args: cobra.NoArgs, RunE: handleList}
var contractsInfoCmd = &cobra.Command{Use: "info <address>", Short: "Show ricardian manifest", Args: cobra.ExactArgs(1), RunE: handleInfo}

func init() {
	deployCmd.Flags().String("wasm", "", "compiled wasm path")
	deployCmd.Flags().String("ric", "", "ricardian manifest JSON (optional)")
	deployCmd.Flags().String("gas", "", "gas limit (default 3M)")

	invokeCmd.Flags().String("method", "", "method name")
	invokeCmd.Flags().String("args", "", "hex‑encoded arg bytes")
	invokeCmd.Flags().Uint64("gas", 200_000, "gas limit")

	debugCmd.Flags().String("method", "", "method name")
	debugCmd.Flags().String("args", "", "hex‑encoded arg bytes")
	debugCmd.Flags().Uint64("gas", 200_000, "gas limit")

	contractsCmd.AddCommand(compileCmd, deployCmd, invokeCmd, debugCmd, contractsListCmd, contractsInfoCmd)
}

// ──────────────────────────────────────────────────────────────────────────────
// Consolidated export
// ──────────────────────────────────────────────────────────────────────────────

var ContractsCmd = contractsCmd

func RegisterContracts(root *cobra.Command) { root.AddCommand(ContractsCmd) }
