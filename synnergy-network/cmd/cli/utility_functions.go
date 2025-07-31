package cli

// cmd/cli/utility_functions.go — misc helpers & conversions exposed via CLI.
// ---------------------------------------------------------------------------
// Pattern:
//   • All cobra *Command* vars grouped at the TOP for quick browsing.
//   • Controller funcs next (pure logic + validation, no side‑effects aside from I/O).
//   • Middleware/helper funcs for env/log wiring.
//   • Consolidated `UtilityRoute` exported at the BOTTOM.
// ---------------------------------------------------------------------------

import (
    "bufio"
    "crypto/sha256"
    "encoding/hex"
    "errors"
    "fmt"
    "io"
    "log"
    "os"
    "strings"

    "golang.org/x/crypto/blake2b"
    "golang.org/x/crypto/ripemd160"
    "golang.org/x/crypto/sha3"

    "github.com/spf13/cobra"

    "synnergy-network/core"
)

// ---------------------------------------------------------------------------
// CLI definitions (TOP section)
// ---------------------------------------------------------------------------

var utilCmd = &cobra.Command{
    Use:   "util",
    Short: "Utility helpers — hashing, conversions, inspection",
}

var hashCmd = &cobra.Command{
    Use:   "hash",
    Short: "Compute a cryptographic hash of the input data",
    Run:   hashHandler,
}

var shortHashCmd = &cobra.Command{
    Use:   "short-hash",
    Short: "Return first4..last4 shorthand of a 32‑byte hash",
    Run:   shortHashHandler,
}

var bytes2AddrCmd = &cobra.Command{
    Use:   "bytes2addr",
    Short: "Convert big‑endian bytes (hex) to a 20‑byte address",
    Run:   bytesToAddressHandler,
}

func init() {
    // hash flags
    hashCmd.Flags().StringP("alg", "a", "sha256", "Hash algorithm: sha256 | keccak256 | ripemd160 | blake2b256")
    hashCmd.Flags().StringP("data", "d", "", "Input data as string – if omitted, reads from STDIN")
    hashCmd.Flags().StringP("file", "f", "", "Path to file to hash (overrides --data)")

    // short‑hash flags
    shortHashCmd.Flags().StringP("hash", "s", "", "32‑byte hash hex string [required]")

    // bytes2addr flags
    bytes2AddrCmd.Flags().StringP("bytes", "b", "", "Hex string representing big‑endian bytes [required]")

    // Register sub‑commands
    utilCmd.AddCommand(hashCmd)
    utilCmd.AddCommand(shortHashCmd)
    utilCmd.AddCommand(bytes2AddrCmd)
}

// ---------------------------------------------------------------------------
// Controller functions
// ---------------------------------------------------------------------------

func hashHandler(cmd *cobra.Command, args []string) {
    alg, _ := cmd.Flags().GetString("alg")
    dataStr, _ := cmd.Flags().GetString("data")
    filePath, _ := cmd.Flags().GetString("file")

    var data []byte
    var err error

    switch {
    case filePath != "":
        data, err = os.ReadFile(filePath)
        bail(err)
    case dataStr != "":
        data = []byte(dataStr)
    default:
        // read from STDIN
        in, err := io.ReadAll(bufio.NewReader(os.Stdin))
        bail(err)
        data = in
    }

    var sum []byte
    switch strings.ToLower(alg) {
    case "sha256":
        v := sha256.Sum256(data)
        sum = v[:]
    case "keccak256":
        h := sha3.NewLegacyKeccak256()
        h.Write(data)
        sum = h.Sum(nil)
    case "ripemd160":
        h := ripemd160.New()
        h.Write(data)
        raw := h.Sum(nil)
        // left‑pad to 32 bytes to align with EVM semantics
        sum = make([]byte, 32)
        copy(sum[32-len(raw):], raw)
    case "blake2b256":
        v := blake2b.Sum256(data)
        sum = v[:]
    default:
        bail(fmt.Errorf("unsupported algorithm: %s", alg))
    }

    fmt.Printf("%x\n", sum)
}

func shortHashHandler(cmd *cobra.Command, args []string) {
    hHex, _ := cmd.Flags().GetString("hash")
    if hHex == "" {
        _ = cmd.Usage()
        bail(errors.New("--hash is required"))
    }
    bytes, err := hex.DecodeString(strings.TrimPrefix(hHex, "0x"))
    bail(err)
    if len(bytes) != 32 {
        bail(fmt.Errorf("hash must be 32 bytes, got %d", len(bytes)))
    }
    var h core.Hash
    copy(h[:], bytes)
    fmt.Println(h.Short())
}

func bytesToAddressHandler(cmd *cobra.Command, args []string) {
    bHex, _ := cmd.Flags().GetString("bytes")
    if bHex == "" {
        _ = cmd.Usage()
        bail(errors.New("--bytes is required"))
    }
    raw, err := hex.DecodeString(strings.TrimPrefix(bHex, "0x"))
    bail(err)
    addr := core.BytesToAddress(raw)
    fmt.Printf("0x%x\n", addr[:])
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func bail(err error) {
    if err != nil {
        log.Fatalf("❌ %v", err)
    }
}

// ---------------------------------------------------------------------------
// Consolidated route export (BOTTOM)
// ---------------------------------------------------------------------------

// UtilityRoute is the single import for root CLI registration.
var UtilityRoute = utilCmd

// ---------------------------------------------------------------------------
// END cmd/cli/utility_functions.go
// ---------------------------------------------------------------------------
