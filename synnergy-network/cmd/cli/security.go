// cmd/cli/security.go – Crypto & security primitives CLI
// -----------------------------------------------------------------------------
// Consolidated under the route “~sec” (aliases: ~security).  Provides thin
// wrappers around the core security primitives exposed via the security daemon
// (newline‑framed JSON‑TCP).
// -----------------------------------------------------------------------------
// Commands
//  sign            – sign a message with a private key (Ed25519/BLS)
//  verify          – verify signature
//  aggregate       – aggregate compressed BLS signatures
//  encrypt         – XChaCha20‑Poly1305 encrypt
//  decrypt         – decrypt
//  merkle          – compute double‑SHA256 Merkle root of leaves
//  dilithium-gen   – generate Dilithium key pair
//  dilithium-sign  – sign message with Dilithium private key
//  dilithium-verify– verify Dilithium signature
//  anomaly-score   – compute anomaly score for a value
// -----------------------------------------------------------------------------
// Environment
//   SECURITY_API_ADDR – host:port of security daemon (default "127.0.0.1:7970")
// -----------------------------------------------------------------------------

package cli

import (
	"bufio"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	core "synnergy-network/core"
)

// -----------------------------------------------------------------------------
// Middleware – framed JSON/TCP client
// -----------------------------------------------------------------------------

type secClient struct {
	conn net.Conn
	rd   *bufio.Reader
}

func newSecClient(ctx context.Context) (*secClient, error) {
	addr := viper.GetString("SECURITY_API_ADDR")
	if addr == "" {
		addr = "127.0.0.1:7970"
	}
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to security daemon at %s: %w", addr, err)
	}
	return &secClient{conn: conn, rd: bufio.NewReader(conn)}, nil
}

func (c *secClient) Close() { _ = c.conn.Close() }

func (c *secClient) writeJSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = c.conn.Write(b)
	return err
}

func (c *secClient) readJSON(v any) error {
	dec := json.NewDecoder(c.rd)
	return dec.Decode(v)
}

// -----------------------------------------------------------------------------
// Controller helpers
// -----------------------------------------------------------------------------

func signRPC(ctx context.Context, algo, keyHex, msgHex string) (string, error) {
	cli, err := newSecClient(ctx)
	if err != nil {
		return "", err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "sign", "algo": algo, "key": keyHex, "msg": msgHex}); err != nil {
		return "", err
	}
	var resp struct {
		Sig   string `json:"sig"`
		Error string `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}
	return resp.Sig, nil
}

func verifyRPC(ctx context.Context, algo, pubHex, msgHex, sigHex string) (bool, error) {
	cli, err := newSecClient(ctx)
	if err != nil {
		return false, err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "verify", "algo": algo, "pub": pubHex, "msg": msgHex, "sig": sigHex}); err != nil {
		return false, err
	}
	var resp struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return false, err
	}
	if resp.Error != "" {
		return false, errors.New(resp.Error)
	}
	return resp.Ok, nil
}

func aggregateRPC(ctx context.Context, sigs []string) (string, error) {
	cli, err := newSecClient(ctx)
	if err != nil {
		return "", err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "aggregate", "sigs": sigs}); err != nil {
		return "", err
	}
	var resp struct {
		Agg   string `json:"agg"`
		Error string `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}
	return resp.Agg, nil
}

func encryptRPC(ctx context.Context, keyHex, msgHex string) (string, error) {
	cli, err := newSecClient(ctx)
	if err != nil {
		return "", err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "encrypt", "key": keyHex, "msg": msgHex}); err != nil {
		return "", err
	}
	var resp struct {
		Blob  string `json:"blob"`
		Error string `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}
	return resp.Blob, nil
}

func decryptRPC(ctx context.Context, keyHex, blobHex string) (string, error) {
	cli, err := newSecClient(ctx)
	if err != nil {
		return "", err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "decrypt", "key": keyHex, "blob": blobHex}); err != nil {
		return "", err
	}
	var resp struct {
		Msg   string `json:"msg"`
		Error string `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}
	return resp.Msg, nil
}

func merkleRPC(ctx context.Context, leaves []string) (string, error) {
	cli, err := newSecClient(ctx)
	if err != nil {
		return "", err
	}
	defer cli.Close()
	if err := cli.writeJSON(map[string]any{"action": "merkle", "leaves": leaves}); err != nil {
		return "", err
	}
	var resp struct {
		Root  string `json:"root"`
		Error string `json:"error,omitempty"`
	}
	if err := cli.readJSON(&resp); err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}
	return resp.Root, nil
}

// -----------------------------------------------------------------------------
// Top‑level Cobra command tree
// -----------------------------------------------------------------------------

var secCmd = &cobra.Command{
	Use:     "~sec",
	Short:   "Security & crypto primitives",
	Aliases: []string{"sec", "security"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cobra.OnInitialize(initSecConfig)
		return nil
	},
}

// sign ------------------------------------------------------------------------
var secSignCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign a message (hex) with a private key (hex)",
	RunE: func(cmd *cobra.Command, args []string) error {
		algo, _ := cmd.Flags().GetString("algo")
		key, _ := cmd.Flags().GetString("key")
		msg, _ := cmd.Flags().GetString("msg")
		if algo == "" || key == "" || msg == "" {
			return errors.New("--algo --key --msg required")
		}
		if _, err := hex.DecodeString(key); err != nil {
			return fmt.Errorf("invalid key hex: %w", err)
		}
		if _, err := hex.DecodeString(msg); err != nil {
			return fmt.Errorf("invalid msg hex: %w", err)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		sig, err := signRPC(ctx, algo, key, msg)
		if err != nil {
			return err
		}
		fmt.Println(sig)
		return nil
	},
}

// verify ----------------------------------------------------------------------
var secVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify signature",
	RunE: func(cmd *cobra.Command, args []string) error {
		algo, _ := cmd.Flags().GetString("algo")
		pub, _ := cmd.Flags().GetString("pub")
		msg, _ := cmd.Flags().GetString("msg")
		sig, _ := cmd.Flags().GetString("sig")
		if algo == "" || pub == "" || msg == "" || sig == "" {
			return errors.New("--algo --pub --msg --sig required")
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		ok, err := verifyRPC(ctx, algo, pub, msg, sig)
		if err != nil {
			return err
		}
		if ok {
			fmt.Println("OK")
		} else {
			fmt.Println("FAIL")
		}
		return nil
	},
}

// aggregate -------------------------------------------------------------------
var aggregateCmd = &cobra.Command{
	Use:   "aggregate [sig1,sig2,…]",
	Short: "Aggregate compressed BLS signatures (hex32 comma list)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sigs := strings.Split(args[0], ",")
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		agg, err := aggregateRPC(ctx, sigs)
		if err != nil {
			return err
		}
		fmt.Println(agg)
		return nil
	},
}

// encrypt ---------------------------------------------------------------------
var encCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt (XChaCha20‑Poly1305) – outputs hex blob",
	RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("key")
		msg, _ := cmd.Flags().GetString("msg")
		if key == "" || msg == "" {
			return errors.New("--key --msg required")
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		blob, err := encryptRPC(ctx, key, msg)
		if err != nil {
			return err
		}
		fmt.Println(blob)
		return nil
	},
}

// decrypt ---------------------------------------------------------------------
var decCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt blob (hex) with key",
	RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("key")
		blob, _ := cmd.Flags().GetString("blob")
		if key == "" || blob == "" {
			return errors.New("--key --blob required")
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
		defer cancel()
		msg, err := decryptRPC(ctx, key, blob)
		if err != nil {
			return err
		}
		fmt.Println(msg)
		return nil
	},
}

// merkle ----------------------------------------------------------------------
var merkleCmd = &cobra.Command{
	Use:   "merkle [leaf1,leaf2,…]",
	Short: "Compute double‑SHA256 Merkle root (hex)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		leaves := strings.Split(args[0], ",")
		ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
		defer cancel()
		root, err := merkleRPC(ctx, leaves)
		if err != nil {
			return err
		}
		fmt.Println(root)
		return nil
	},
}

// dilithium-gen ---------------------------------------------------------------
var dilGenCmd = &cobra.Command{
	Use:   "dilithium-gen",
	Short: "Generate Dilithium3 key pair",
	RunE: func(cmd *cobra.Command, args []string) error {
		pub, priv, err := core.DilithiumKeypair()
		if err != nil {
			return err
		}
		fmt.Printf("pub:%x\npriv:%x\n", pub, priv)
		return nil
	},
}

// dilithium-sign --------------------------------------------------------------
var dilSignCmd = &cobra.Command{
	Use:   "dilithium-sign",
	Short: "Sign message (hex) with Dilithium private key",
	RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("key")
		msg, _ := cmd.Flags().GetString("msg")
		if key == "" || msg == "" {
			return errors.New("--key --msg required")
		}
		priv, err := hex.DecodeString(key)
		if err != nil {
			return fmt.Errorf("bad key: %w", err)
		}
		m, err := hex.DecodeString(msg)
		if err != nil {
			return fmt.Errorf("bad msg: %w", err)
		}
		sig, err := core.DilithiumSign(priv, m)
		if err != nil {
			return err
		}
		fmt.Println(hex.EncodeToString(sig))
		return nil
	},
}

// dilithium-verify ------------------------------------------------------------
var dilVerifyCmd = &cobra.Command{
	Use:   "dilithium-verify",
	Short: "Verify Dilithium signature",
	RunE: func(cmd *cobra.Command, args []string) error {
		pubHex, _ := cmd.Flags().GetString("pub")
		msgHex, _ := cmd.Flags().GetString("msg")
		sigHex, _ := cmd.Flags().GetString("sig")
		if pubHex == "" || msgHex == "" || sigHex == "" {
			return errors.New("--pub --msg --sig required")
		}
		pub, err := hex.DecodeString(pubHex)
		if err != nil {
			return fmt.Errorf("bad pub: %w", err)
		}
		msg, err := hex.DecodeString(msgHex)
		if err != nil {
			return fmt.Errorf("bad msg: %w", err)
		}
		sig, err := hex.DecodeString(sigHex)
		if err != nil {
			return fmt.Errorf("bad sig: %w", err)
		}
		ok, err := core.DilithiumVerify(pub, msg, sig)
		if err != nil {
			return err
		}
		if ok {
			fmt.Println("OK")
		} else {
			fmt.Println("FAIL")
		}
		return nil
	},
}

// anomaly-score ---------------------------------------------------------------
var anomalyCmd = &cobra.Command{
	Use:   "anomaly-score",
	Short: "Compute anomaly z-score for value based on comma-separated data",
	RunE: func(cmd *cobra.Command, args []string) error {
		dataStr, _ := cmd.Flags().GetString("data")
		val, _ := cmd.Flags().GetFloat64("value")
		if dataStr == "" {
			return errors.New("--data required")
		}
		det := core.NewAnomalyDetector()
		for _, part := range strings.Split(dataStr, ",") {
			f, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
			if err != nil {
				return fmt.Errorf("invalid data %q: %w", part, err)
			}
			det.Update(f)
		}
		score := det.Score(val)
		fmt.Printf("%f\n", score)
		return nil
	},
}

// -----------------------------------------------------------------------------
// init – config & route wiring
// -----------------------------------------------------------------------------

func initSecConfig() {
	viper.SetEnvPrefix("synnergy")
	viper.AutomaticEnv()

	cfg := viper.GetString("config")
	if cfg != "" {
		viper.SetConfigFile(cfg)
	} else {
		viper.SetConfigName("synnergy")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config/synnergy")
	}
	_ = viper.ReadInConfig()

	viper.SetDefault("SECURITY_API_ADDR", "127.0.0.1:7970")
}

func init() {
	// sign flags
	secSignCmd.Flags().String("algo", "", "algo: ed25519|bls")
	secSignCmd.Flags().String("key", "", "private key hex / compressed")
	secSignCmd.Flags().String("msg", "", "message hex")

	// verify flags
	secVerifyCmd.Flags().String("algo", "", "algo: ed25519|bls")
	secVerifyCmd.Flags().String("pub", "", "public key hex / compressed")
	secVerifyCmd.Flags().String("msg", "", "message hex")
	secVerifyCmd.Flags().String("sig", "", "signature hex")

	// encrypt / decrypt flags
	encCmd.Flags().String("key", "", "32‑byte key hex")
	encCmd.Flags().String("msg", "", "plaintext hex")
	decCmd.Flags().String("key", "", "32‑byte key hex")
	decCmd.Flags().String("blob", "", "ciphertext blob hex")

	// dilithium-sign flags
	dilSignCmd.Flags().String("key", "", "private key hex")
	dilSignCmd.Flags().String("msg", "", "message hex")

	// dilithium-verify flags
	dilVerifyCmd.Flags().String("pub", "", "public key hex")
	dilVerifyCmd.Flags().String("msg", "", "message hex")
	dilVerifyCmd.Flags().String("sig", "", "signature hex")

	// anomaly-score flags
	anomalyCmd.Flags().String("data", "", "comma separated floats")
	anomalyCmd.Flags().Float64("value", 0, "value to score")

	// register sub‑commands
	secCmd.AddCommand(secSignCmd)
	secCmd.AddCommand(secVerifyCmd)
	secCmd.AddCommand(aggregateCmd)
	secCmd.AddCommand(encCmd)
	secCmd.AddCommand(decCmd)
	secCmd.AddCommand(merkleCmd)
	secCmd.AddCommand(dilGenCmd)
	secCmd.AddCommand(dilSignCmd)
	secCmd.AddCommand(dilVerifyCmd)
	secCmd.AddCommand(anomalyCmd)
}

// NewSecurityCommand exposes the consolidated command tree.
func NewSecurityCommand() *cobra.Command { return secCmd }
