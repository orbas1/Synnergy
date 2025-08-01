package cli

// ──────────────────────────────────────────────────────────────────────────────
// Synthron Transactions CLI – craft, sign, verify & submit network txs
//
// The CLI mirrors the structure of other modules (coin, consensus):
//   • Primary command objects declared first for quick scanning.
//   • Shared middleware initialises required services once (txLedger, p2p, keys,
//     authority, txpool, gas‑calculator).
//   • Controllers implement the business logic for each route.
//   • All routes are consolidated at the bottom and exported as `TransactionsCmd`.
//
// Environment requirements (add to .env or orchestration layer):
//   • LEDGER_PATH         – Bolt/Badger DB for state (shared).
//   • KEYSTORE_PATH       – Directory with PEM keys (sec svc).
//   • P2P_PORT            – TCP port for p2p (default 30333).
//   • P2P_BOOTNODES       – Comma‑separated multiaddr list.
//   • AUTH_DB_PATH        – Path to authority DB (SQLite, Bolt, …).
//   • LOG_LEVEL           – trace|debug|info|warn|error (default info).
//
// Example usage once wired into root CLI:
//   ~tx ~create  --to 0xabc… --value 100000 --gas 21000 > raw.json
//   ~tx ~sign    --in raw.json --out signed.json
//   ~tx ~submit  --in signed.json
//   ~tx ~verify  --in signed.json
//   ~tx ~pool
//
// ──────────────────────────────────────────────────────────────────────────────

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"synnergy-network/core"
)

// ──────────────────────────────────────────────────────────────────────────────
// Globals & middleware
// ──────────────────────────────────────────────────────────────────────────────

var (
	txPoolSvc *core.TxPool
	secSvc    *keyService
	txLogger  = logrus.StandardLogger()
	txLedger  *core.Ledger

	// protects one‑time init within PersistentPreRunE
	initOnce sync.Once
)

// keyService provides minimal PEM key loading for signing transactions.
type keyService struct{ path string }

func newKeyService(p string) *keyService { return &keyService{path: p} }

func (ks *keyService) LoadKey(name string) (interface{}, error) {
	b, err := ioutil.ReadFile(filepath.Join(ks.path, name))
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM")
	}
	return x509.ParsePKCS8PrivateKey(block.Bytes)
}

func initTxMiddleware(cmd *cobra.Command, _ []string) error {
	var retErr error
	initOnce.Do(func() {
		// 1. Load env
		_ = godotenv.Load()

		// 2. Logger level
		lvlStr := os.Getenv("LOG_LEVEL")
		if lvlStr == "" {
			lvlStr = "info"
		}
		lvl, err := logrus.ParseLevel(lvlStr)
		if err != nil {
			retErr = fmt.Errorf("invalid LOG_LEVEL: %w", err)
			return
		}
		txLogger.SetLevel(lvl)

		// 3. Ledger
		lp := os.Getenv("LEDGER_PATH")
		if lp == "" {
			retErr = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		txLedger, err = core.OpenLedger(lp)
		if err != nil {
			retErr = fmt.Errorf("open txLedger: %w", err)
			return
		}

		// 4. P2P network (light footprint)
		port := os.Getenv("P2P_PORT")
		if port == "" {
			port = "30333"
		}
		p2pSvc, err := core.NewNode(core.Config{
			ListenAddr:     fmt.Sprintf(":%s", port),
			BootstrapPeers: strings.Split(os.Getenv("P2P_BOOTNODES"), ","),
			DiscoveryTag:   "cli",
		})
		if err != nil {
			retErr = fmt.Errorf("init p2p: %w", err)
			return
		}

		// 5. Security / keystore
		ks := os.Getenv("KEYSTORE_PATH")
		if ks == "" {
			retErr = fmt.Errorf("KEYSTORE_PATH not set")
			return
		}
		secSvc = newKeyService(ks)

		// 6. Authority (for tx reversal checks)
		authSvc := core.NewAuthoritySet(txLogger, txLedger)

		// 7. Gas calculator – placeholder flat gas until economics stabilises
		gasCalc := core.NewFlatGasCalculator(10) // 10 wei per gas unit

		// 8. TxPool
		txPoolSvc = core.NewTxPool(nil, txLedger, authSvc, gasCalc, p2pSvc, 0)

		// background processor
		go txPoolSvc.Run(context.Background())
	})
	return retErr
}

// ──────────────────────────────────────────────────────────────────────────────
// Controller helpers
// ──────────────────────────────────────────────────────────────────────────────

type txCreateFlags struct {
	to       string
	value    uint64
	gasLimit uint64
	gasPrice uint64
	nonce    uint64
	payload  string
	txType   string
	output   string
}

func txHandleCreate(cmd *cobra.Command, _ []string) error {
	flags := cmd.Context().Value("flags").(txCreateFlags)

	var toAddr core.Address
	if flags.to != "" {
		b, err := hex.DecodeString(strings.TrimPrefix(flags.to, "0x"))
		if err != nil || len(b) != len(toAddr) {
			return fmt.Errorf("invalid --to address")
		}
		copy(toAddr[:], b)
	}

	var t core.TxType
	switch strings.ToLower(flags.txType) {
	case "payment", "pay":
		t = core.TxPayment
	case "call", "contract":
		t = core.TxContractCall
	case "reversal":
		t = core.TxReversal
	default:
		return fmt.Errorf("unknown --type (payment|call|reversal)")
	}

	tx := &core.Transaction{
		Type:      t,
		To:        toAddr,
		Value:     flags.value,
		GasLimit:  flags.gasLimit,
		GasPrice:  flags.gasPrice,
		Nonce:     flags.nonce,
		Payload:   []byte(flags.payload),
		Timestamp: time.Now().UnixMilli(),
	}
	tx.HashTx()

	jsonBytes, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		return err
	}

	if flags.output != "" {
		if err := ioutil.WriteFile(flags.output, jsonBytes, 0o600); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "unsigned tx written to %s\n", flags.output)
	} else {
		cmd.OutOrStdout().Write(jsonBytes)
		fmt.Fprintln(cmd.OutOrStdout())
	}
	return nil
}

type txSignFlags struct {
	input  string
	output string
	key    string
}

func txHandleSign(cmd *cobra.Command, _ []string) error {
	flags := cmd.Context().Value("sflags").(txSignFlags)

	raw, err := ioutil.ReadFile(flags.input)
	if err != nil {
		return err
	}

	var tx core.Transaction
	if err := json.Unmarshal(raw, &tx); err != nil {
		return fmt.Errorf("decode tx: %w", err)
	}

	// resolve key (alias=filename in keystore)
	privKey, err := secSvc.LoadKey(flags.key)
	if err != nil {
		return err
	}

	if err := tx.Sign(privKey.(*ecdsa.PrivateKey)); err != nil {
		return fmt.Errorf("sign: %w", err)
	}
	jsonBytes, _ := json.MarshalIndent(&tx, "", "  ")

	if flags.output != "" {
		if err := ioutil.WriteFile(flags.output, jsonBytes, 0o600); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "signed tx written to %s\n", flags.output)
	} else {
		cmd.OutOrStdout().Write(jsonBytes)
		fmt.Fprintln(cmd.OutOrStdout())
	}
	return nil
}

type verifyFlags struct{ input string }

func handleVerify(cmd *cobra.Command, _ []string) error {
	in := cmd.Context().Value("vflags").(verifyFlags).input
	raw, err := ioutil.ReadFile(in)
	if err != nil {
		return err
	}

	var tx core.Transaction
	if err := json.Unmarshal(raw, &tx); err != nil {
		return err
	}
	if err := tx.VerifySig(); err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "signature OK ✔")
	return nil
}

type submitFlags struct{ input string }

func handleSubmit(cmd *cobra.Command, _ []string) error {
	in := cmd.Context().Value("subflags").(submitFlags).input
	raw, err := ioutil.ReadFile(in)
	if err != nil {
		return err
	}

	var tx core.Transaction
	if err := json.Unmarshal(raw, &tx); err != nil {
		return err
	}
	if err := txPoolSvc.AddTx(&tx); err != nil {
		return fmt.Errorf("pool reject: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "tx %s accepted\n", tx.IDHex())
	return nil
}

func handlePool(cmd *cobra.Command, _ []string) error {
	list := txPoolSvc.Snapshot()
	for _, tx := range list {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", tx.IDHex())
	}
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Cobra commands (primary – declared before init())
// ──────────────────────────────────────────────────────────────────────────────

var txCmd = &cobra.Command{
	Use:               "tx",
	Short:             "Create, sign, verify and submit Synthron transactions",
	PersistentPreRunE: initTxMiddleware,
}

// create
var txCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Craft an unsigned transaction JSON",
	Args:  cobra.NoArgs,
	RunE:  txHandleCreate,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		cf := txCreateFlags{}
		cf.to, _ = cmd.Flags().GetString("to")
		cf.value, _ = cmd.Flags().GetUint64("value")
		cf.gasLimit, _ = cmd.Flags().GetUint64("gas")
		cf.gasPrice, _ = cmd.Flags().GetUint64("price")
		cf.nonce, _ = cmd.Flags().GetUint64("nonce")
		cf.payload, _ = cmd.Flags().GetString("payload")
		cf.txType, _ = cmd.Flags().GetString("type")
		cf.output, _ = cmd.Flags().GetString("out")
		ctx := context.WithValue(cmd.Context(), "flags", cf)
		cmd.SetContext(ctx)
		return nil
	},
}

// sign
var txSignCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign a JSON transaction with a private key from keystore",
	Args:  cobra.NoArgs,
	RunE:  txHandleSign,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		sf := txSignFlags{}
		sf.input, _ = cmd.Flags().GetString("in")
		sf.output, _ = cmd.Flags().GetString("out")
		sf.key, _ = cmd.Flags().GetString("key")
		ctx := context.WithValue(cmd.Context(), "sflags", sf)
		cmd.SetContext(ctx)
		return nil
	},
}

// verify
var txVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify a signed transaction JSON",
	Args:  cobra.NoArgs,
	RunE:  handleVerify,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		vf := verifyFlags{}
		vf.input, _ = cmd.Flags().GetString("in")
		ctx := context.WithValue(cmd.Context(), "vflags", vf)
		cmd.SetContext(ctx)
		return nil
	},
}

// submit
var txSubmitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Add a signed transaction to the mem‑pool & broadcast",
	Args:  cobra.NoArgs,
	RunE:  handleSubmit,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		sf := submitFlags{}
		sf.input, _ = cmd.Flags().GetString("in")
		ctx := context.WithValue(cmd.Context(), "subflags", sf)
		cmd.SetContext(ctx)
		return nil
	},
}

// pool
var txPoolCmd = &cobra.Command{
	Use:   "pool",
	Short: "List pending pool transaction hashes",
	Args:  cobra.NoArgs,
	RunE:  handlePool,
}

func init() {
	// create flags
	txCreateCmd.Flags().String("to", "", "hex recipient address (0x…)")
	txCreateCmd.Flags().Uint64("value", 0, "value in wei")
	txCreateCmd.Flags().Uint64("gas", 21_000, "gas limit")
	txCreateCmd.Flags().Uint64("price", 1, "gas price in wei")
	txCreateCmd.Flags().Uint64("nonce", 0, "transaction nonce")
	txCreateCmd.Flags().String("payload", "", "optional input data (hex/string)")
	txCreateCmd.Flags().String("type", "payment", "payment|call|reversal")
	txCreateCmd.Flags().String("out", "", "output file path (defaults to stdout)")

	// sign flags
	txSignCmd.Flags().String("in", "", "input JSON file")
	txSignCmd.MarkFlagRequired("in")
	txSignCmd.Flags().String("out", "", "output file (defaults stdout)")
	txSignCmd.Flags().String("key", "node", "keystore key alias")

	// verify
	txVerifyCmd.Flags().String("in", "", "input JSON file")
	txVerifyCmd.MarkFlagRequired("in")

	// submit
	txSubmitCmd.Flags().String("in", "", "signed JSON file")
	txSubmitCmd.MarkFlagRequired("in")

	// assemble tree
	txCmd.AddCommand(txCreateCmd)
	txCmd.AddCommand(txSignCmd)
	txCmd.AddCommand(txVerifyCmd)
	txCmd.AddCommand(txSubmitCmd)
	txCmd.AddCommand(txPoolCmd)
}

// ──────────────────────────────────────────────────────────────────────────────
// Consolidated route export
// ──────────────────────────────────────────────────────────────────────────────

var TransactionsCmd = txCmd

func RegisterTransactions(root *cobra.Command) { root.AddCommand(TransactionsCmd) }
