package cli

// ──────────────────────────────────────────────────────────────────────────────
// Synthron Wallet CLI – HD key management & transaction signing
//
// Root command:  `wallet`
// Sub‑routes:
//   create   – generate fresh 12/24‑word mnemonic + save encrypted wallet file
//   import   – import mnemonic and create wallet file
//   address  – derive address at account/index
//   sign     – sign a transaction JSON using derived key
//
// Wallet file layout (JSON, encrypted with PBKDF2‑AES‑256‑GCM):
//   {
//     "seed": <hex>,
//     "salt": <hex>,
//     "nonce": <hex>,
//     "cipher": <hex>
//   }
//
// Env vars:
//   LOG_LEVEL          – trace|debug|info|warn|error (default info)
//
// ──────────────────────────────────────────────────────────────────────────────

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "sync"

    "github.com/joho/godotenv"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    "golang.org/x/crypto/pbkdf2"

    "synnergy-network/core"
)

// ──────────────────────────────────────────────────────────────────────────────
// Globals & middleware
// ──────────────────────────────────────────────────────────────────────────────

var (
    logger = logrus.StandardLogger()
    once   sync.Once
)

func initWalletMiddleware(cmd *cobra.Command, _ []string) error {
    var err error
    once.Do(func() {
        _ = godotenv.Load()
        lvl := os.Getenv("LOG_LEVEL")
        if lvl == "" { lvl = "info" }
        l, e := logrus.ParseLevel(lvl)
        if e != nil { err = e; return }
        logger.SetLevel(l)
        core.SetWalletLogger(logger)
    })
    return err
}

// ──────────────────────────────────────────────────────────────────────────────
// Helpers – encrypt / decrypt wallet file
// ──────────────────────────────────────────────────────────────────────────────

type keystore struct {
    Seed   string `json:"seed"`
    Salt   string `json:"salt"`
    Nonce  string `json:"nonce"`
    Cipher string `json:"cipher"`
}

func deriveKey(password string, salt []byte) []byte {
    return pbkdf2.Key([]byte(password), salt, 150_000, 32, sha256.New)
}

func encryptSeed(seed []byte, password string) (*keystore, error) {
    salt := make([]byte, 16); if _, err := rand.Read(salt); err != nil { return nil, err }
    key := deriveKey(password, salt)
    block, err := aes.NewCipher(key); if err != nil { return nil, err }
    gcm, err := cipher.NewGCM(block); if err != nil { return nil, err }
    nonce := make([]byte, gcm.NonceSize()); if _, err := rand.Read(nonce); err != nil { return nil, err }
    cipherText := gcm.Seal(nil, nonce, seed, nil)
    return &keystore{
        Salt:   hex.EncodeToString(salt),
        Nonce:  hex.EncodeToString(nonce),
        Cipher: hex.EncodeToString(cipherText),
    }, nil
}

func decryptSeed(ks *keystore, password string) ([]byte, error) {
    salt, _ := hex.DecodeString(ks.Salt)
    nonce, _ := hex.DecodeString(ks.Nonce)
    cipherText, _ := hex.DecodeString(ks.Cipher)
    key := deriveKey(password, salt)
    block, err := aes.NewCipher(key); if err != nil { return nil, err }
    gcm, err := cipher.NewGCM(block); if err != nil { return nil, err }
    return gcm.Open(nil, nonce, cipherText, nil)
}

// ──────────────────────────────────────────────────────────────────────────────
// Controller logic
// ──────────────────────────────────────────────────────────────────────────────

type createFlags struct {
    bits int
    out  string
    pwd  string
}

type importFlags struct {
    mnemonic   string
    passphrase string
    pwd        string
    out        string
}

type addrFlags struct {
    wallet string
    pwd    string
    acct   uint32
    idx    uint32
}

type signFlags struct {
    wallet   string
    pwd      string
    acct     uint32
    idx      uint32
    txIn     string
    txOut    string
    gasPrice uint64
}

func handleCreate(cmd *cobra.Command, _ []string) error {
    cf := cmd.Context().Value("cflags").(createFlags)
    w, mnemonic, err := core.NewRandomWallet(cf.bits)
    if err != nil { return err }

    ks, err := encryptSeed(w.Seed(), cf.pwd)
    if err != nil { return err }
    data, _ := json.MarshalIndent(ks, "", "  ")

    if cf.out != "" {
        if err := ioutil.WriteFile(cf.out, data, 0o600); err != nil { return err }
        fmt.Fprintf(cmd.OutOrStdout(), "wallet saved to %s\n", cf.out)
    } else {
        cmd.OutOrStdout().Write(data)
        fmt.Fprintln(cmd.OutOrStdout())
    }
    fmt.Fprintf(cmd.OutOrStdout(), "mnemonic (WRITE IT DOWN): %s\n", mnemonic)
    return nil
}

func handleImport(cmd *cobra.Command, _ []string) error {
    f := cmd.Context().Value("iflags").(importFlags)
    w, err := core.WalletFromMnemonic(f.mnemonic, f.passphrase)
    if err != nil { return err }
    ks, err := encryptSeed(w.Seed(), f.pwd)
    if err != nil { return err }
    data, _ := json.MarshalIndent(ks, "", "  ")
    if f.out != "" {
        if err := ioutil.WriteFile(f.out, data, 0o600); err != nil { return err }
        fmt.Fprintf(cmd.OutOrStdout(), "wallet saved to %s\n", f.out)
    } else {
        cmd.OutOrStdout().Write(data)
        fmt.Fprintln(cmd.OutOrStdout())
    }
    return nil
}

func loadWallet(path, pwd string) (*core.HDWallet, error) {
    raw, err := ioutil.ReadFile(filepath.Clean(path)); if err != nil { return nil, err }
    var ks keystore; if err := json.Unmarshal(raw, &ks); err != nil { return nil, err }
    seed, err := decryptSeed(&ks, pwd); if err != nil { return nil, err }
    return core.NewHDWalletFromSeed(seed, logger)
}

func handleAddress(cmd *cobra.Command, _ []string) error {
    af := cmd.Context().Value("aflags").(addrFlags)
    w, err := loadWallet(af.wallet, af.pwd); if err != nil { return err }
    addr, err := w.NewAddress(af.acct, af.idx); if err != nil { return err }
    fmt.Fprintln(cmd.OutOrStdout(), addr.Hex())
    return nil
}

func handleSign(cmd *cobra.Command, _ []string) error {
    sf := cmd.Context().Value("sflags").(signFlags)
    w, err := loadWallet(sf.wallet, sf.pwd); if err != nil { return err }
    raw, err := ioutil.ReadFile(sf.txIn); if err != nil { return err }
    var tx core.Transaction
    if err := json.Unmarshal(raw, &tx); err != nil { return err }
    if err := w.SignTx(&tx, sf.acct, sf.idx, sf.gasPrice); err != nil { return err }
    out, _ := json.MarshalIndent(&tx, "", "  ")
    if sf.txOut != "" {
        if err := ioutil.WriteFile(sf.txOut, out, 0o600); err != nil { return err }
        fmt.Fprintf(cmd.OutOrStdout(), "signed tx written to %s\n", sf.txOut)
    } else {
        cmd.OutOrStdout().Write(out)
        fmt.Fprintln(cmd.OutOrStdout())
    }
    return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Cobra command tree
// ──────────────────────────────────────────────────────────────────────────────

var walletCmd = &cobra.Command{
    Use:               "wallet",
    Short:             "HD wallet management & tx signing",
    PersistentPreRunE: initWalletMiddleware,
}

var createCmd = &cobra.Command{
    Use:  "create",
    Args: cobra.NoArgs,
    Short: "Generate a new wallet & mnemonic",
    RunE:  handleCreate,
    PreRunE: func(cmd *cobra.Command, args []string) error {
        cf := createFlags{}
        cf.bits, _ = cmd.Flags().GetInt("bits")
        cf.out, _ = cmd.Flags().GetString("out")
        cf.pwd, _ = cmd.Flags().GetString("password")
        if cf.pwd == "" { return errors.New("--password required") }
        ctx := context.WithValue(cmd.Context(), "cflags", cf)
        cmd.SetContext(ctx)
        return nil
    },
}

var importCmd = &cobra.Command{
    Use:   "import",
    Short: "Import existing mnemonic",
    Args:  cobra.NoArgs,
    RunE:  handleImport,
    PreRunE: func(cmd *cobra.Command, args []string) error {
        inf := importFlags{}
        inf.mnemonic, _ = cmd.Flags().GetString("mnemonic")
        inf.passphrase, _ = cmd.Flags().GetString("passphrase")
        inf.out, _ = cmd.Flags().GetString("out")
        inf.pwd, _ = cmd.Flags().GetString("password")
        if inf.mnemonic == "" || inf.pwd == "" { return errors.New("--mnemonic and --password required") }
        ctx := context.WithValue(cmd.Context(), "iflags", inf)
        cmd.SetContext(ctx)
        return nil
    },
}

var addressCmd = &cobra.Command{
    Use:   "address",
    Short: "Derive address",
    Args:  cobra.NoArgs,
    RunE:  handleAddress,
    PreRunE: func(cmd *cobra.Command, args []string) error {
        af := addrFlags{}
        af.wallet, _ = cmd.Flags().GetString("wallet")
        af.pwd, _ = cmd.Flags().GetString("password")
        af.acct, _ = cmd.Flags().GetUint32("account")
        af.idx, _ = cmd.Flags().GetUint32("index")
        if af.wallet == "" || af.pwd == "" { return errors.New("--wallet and --password required") }
        ctx := context.WithValue(cmd.Context(), "aflags", af)
        cmd.SetContext(ctx)
        return nil
    },
}

var signCmd = &cobra.Command{
    Use:   "sign",
    Short: "Sign a transaction JSON",
    Args:  cobra.NoArgs,
    RunE:  handleSign,
    PreRunE: func(cmd *cobra.Command, args []string) error {
        sf := signFlags{}
        sf.wallet, _ = cmd.Flags().GetString("wallet")
        sf.pwd, _ = cmd.Flags().GetString("password")
        sf.acct, _ = cmd.Flags().GetUint32("account")
        sf.idx, _ = cmd.Flags().GetUint32("index")
        sf.txIn, _ = cmd.Flags().GetString("in")
        sf.txOut, _ = cmd.Flags().GetString("out")
        sf.gasPrice, _ = cmd.Flags().GetUint64("gas")
        if sf.wallet == "" || sf.pwd == "" || sf.txIn == "" {
            return errors.New("--wallet, --password, --in required")
        }
        ctx := context.WithValue(cmd.Context(), "sflags", sf)
        cmd.SetContext(ctx)
        return nil
    },
}

func init() {
    // create flags
    createCmd.Flags().Int("bits", 128, "entropy bits (128|256)")
    createCmd.Flags().String("out", "", "output wallet file")
    createCmd.Flags().String("password", "", "encryption password")

    // import flags
    importCmd.Flags().String("mnemonic", "", "bip39 words")
    importCmd.Flags().String("passphrase", "", "optional bip39 passphrase")
    importCmd.Flags().String("password", "", "encryption password")
    importCmd.Flags().String("out", "", "output wallet file")

    // address flags
    addressCmd.Flags().String("wallet", "", "wallet file")
    addressCmd.Flags().String("password", "", "wallet password")
    addressCmd.Flags().Uint32("account", 0, "account # (hardened)")
    addressCmd.Flags().Uint32("index", 0, "index # (hardened)")

    // sign flags
    signCmd.Flags().String("wallet", "", "wallet file")
    signCmd.Flags().String("password", "", "wallet password")
    signCmd.Flags().Uint32("account", 0, "account #")
    signCmd.Flags().Uint32("index", 0, "index #")
    signCmd.Flags().String("in", "", "unsigned tx JSON path")
    signCmd.Flags().String("out", "", "output signed tx path (stdout if empty)")
    signCmd.Flags().Uint64("gas", 0, "override gas price (wei)")

    walletCmd.AddCommand(createCmd, importCmd, addressCmd, signCmd)
}

// ──────────────────────────────────────────────────────────────────────────────
// Consolidated export
// ──────────────────────────────────────────────────────────────────────────────

var WalletCmd = walletCmd

func RegisterWallet(root *cobra.Command) { root.AddCommand(WalletCmd) }
