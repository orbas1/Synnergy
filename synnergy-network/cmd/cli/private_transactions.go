package cli

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"synnergy-network/core"
)

var (
	ptPool   *core.TxPool
	ptLedger *core.Ledger
	ptLogger = logrus.StandardLogger()
	ptOnce   sync.Once
)

func ptInit(cmd *cobra.Command, _ []string) error {
	var err error
	ptOnce.Do(func() {
		_ = godotenv.Load()
		lvl := os.Getenv("LOG_LEVEL")
		if lvl == "" {
			lvl = "info"
		}
		lv, e := logrus.ParseLevel(lvl)
		if e != nil {
			err = e
			return
		}
		ptLogger.SetLevel(lv)
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		ptLedger, e = core.OpenLedger(path)
		if e != nil {
			err = e
			return
		}
		auth := core.NewAuthoritySet(ptLogger, ptLedger)
		gas := core.NewFlatGasCalculator(10)
		netSvc, nerr := core.NewNode(core.Config{ListenAddr: ":0", DiscoveryTag: "privtx"})
		if nerr != nil {
			err = nerr
			return
		}
		ptPool = core.NewTxPool(nil, ptLedger, auth, gas, netSvc, 0)
		go ptPool.Run(context.Background())
	})
	return err
}

func handlePTEncrypt(cmd *cobra.Command, _ []string) error {
	payload, _ := cmd.Flags().GetString("payload")
	keyStr, _ := cmd.Flags().GetString("key")
	key, err := hex.DecodeString(strings.TrimPrefix(keyStr, "0x"))
	if err != nil {
		return err
	}
	tx := &core.Transaction{Payload: []byte(payload)}
	if err := core.EncryptTxPayload(tx, key); err != nil {
		return err
	}
	hexEnc, _ := core.EncodeEncryptedHex(tx)
	fmt.Fprintln(cmd.OutOrStdout(), hexEnc)
	return nil
}

func handlePTDecrypt(cmd *cobra.Command, _ []string) error {
	cipherStr, _ := cmd.Flags().GetString("cipher")
	keyStr, _ := cmd.Flags().GetString("key")
	key, err := hex.DecodeString(strings.TrimPrefix(keyStr, "0x"))
	if err != nil {
		return err
	}
	data, err := hex.DecodeString(strings.TrimPrefix(cipherStr, "0x"))
	if err != nil {
		return err
	}
	tx := &core.Transaction{EncryptedPayload: data, Private: true}
	plain, err := core.DecryptTxPayload(tx, key)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(plain))
	return nil
}

func handlePTSend(cmd *cobra.Command, args []string) error {
	raw, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}
	var tx core.Transaction
	if err := json.Unmarshal(raw, &tx); err != nil {
		return err
	}
	if err := core.SubmitPrivateTx(ptPool, &tx); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "tx %s submitted\n", tx.IDHex())
	return nil
}

var privTxCmd = &cobra.Command{
	Use:               "private_tx",
	Short:             "Utilities for private transactions",
	PersistentPreRunE: ptInit,
}

var ptEncryptCmd = &cobra.Command{Use: "encrypt", Short: "Encrypt payload", RunE: handlePTEncrypt}
var ptDecryptCmd = &cobra.Command{Use: "decrypt", Short: "Decrypt payload", RunE: handlePTDecrypt}
var ptSendCmd = &cobra.Command{Use: "send <file>", Short: "Submit private transaction", Args: cobra.ExactArgs(1), RunE: handlePTSend}

func init() {
	ptEncryptCmd.Flags().String("payload", "", "plaintext data")
	ptEncryptCmd.Flags().String("key", "", "hex encryption key")
	ptEncryptCmd.MarkFlagRequired("payload")
	ptEncryptCmd.MarkFlagRequired("key")

	ptDecryptCmd.Flags().String("cipher", "", "hex ciphertext")
	ptDecryptCmd.Flags().String("key", "", "hex encryption key")
	ptDecryptCmd.MarkFlagRequired("cipher")
	ptDecryptCmd.MarkFlagRequired("key")

	privTxCmd.AddCommand(ptEncryptCmd, ptDecryptCmd, ptSendCmd)
}

var PrivateTxCmd = privTxCmd

func RegisterPrivateTx(root *cobra.Command) { root.AddCommand(PrivateTxCmd) }
