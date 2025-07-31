package cli

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"synnergy-network/core"
)

//-------------------------------------------------------------------------
// CLI‑level helpers & middleware
//-------------------------------------------------------------------------

var (
	ledgerPath string
	engine     *core.ChannelEngine
)

func initMiddleware(cmd *cobra.Command, args []string) {
	// Load .env if present
	_ = godotenv.Load()

	// CLI flag overrides env
	if lp, _ := cmd.Flags().GetString("ledger"); lp != "" {
		ledgerPath = lp
	} else if env := os.Getenv("LEDGER_PATH"); env != "" {
		ledgerPath = env
	} else {
		exe, _ := os.Executable()
		ledgerPath = filepath.Join(filepath.Dir(exe), "state.db")
	}

	// init ledger + engine singleton (idempotent)
	led, err := ledger.NewBadgerLedger(ledgerPath)
	if err != nil {
		log.Fatalf("failed to open ledger at %s: %v", ledgerPath, err)
	}
	core.InitStateChannels(led)
	engine = core.Channels()
}

//-------------------------------------------------------------------------
// Controller helpers
//-------------------------------------------------------------------------

func parseChannelAddress(hexStr string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(hexStr)
	if err != nil || len(b) != len(a) {
		return a, errors.New("address must be 20‑byte hex")
	}
	copy(a[:], b)
	return a, nil
}

func parseTokenID(hexStr string) (core.TokenID, error) {
	var id core.TokenID
	b, err := hex.DecodeString(hexStr)
	if err != nil || len(b) != len(id) {
		return id, errors.New("token ID must be 16‑byte hex")
	}
	copy(id[:], b)
	return id, nil
}

//-------------------------------------------------------------------------
// Command handlers (controller layer)
//-------------------------------------------------------------------------

func openChannelHandler(cmd *cobra.Command, args []string) {
	required := []string{"partyA", "partyB", "token", "amountA", "amountB"}
	for _, r := range required {
		if f := cmd.Flag(r); f == nil || f.Value.String() == "" {
			_ = cmd.Usage()
			log.Fatalf("missing required flag --%s", r)
		}
	}

	aHex, _ := cmd.Flags().GetString("partyA")
	bHex, _ := cmd.Flags().GetString("partyB")
	tokHex, _ := cmd.Flags().GetString("token")
	amountAStr, _ := cmd.Flags().GetString("amountA")
	amountBStr, _ := cmd.Flags().GetString("amountB")
	nonce, _ := cmd.Flags().GetUint64("nonce")

	a, err := parseChannelAddress(aHex)
	channelBail(err)
	b, err := parseChannelAddress(bHex)
	channelBail(err)
	t, err := parseTokenID(tokHex)
	channelBail(err)

	amountA, err := strconv.ParseUint(amountAStr, 10, 64)
	channelBail(err)
	amountB, err := strconv.ParseUint(amountBStr, 10, 64)
	channelBail(err)

	id, err := engine.OpenChannel(a, b, t, amountA, amountB, nonce)
	channelBail(err)

	fmt.Printf("✅ Channel opened: %x\n", id)
}

func initiateCloseHandler(cmd *cobra.Command, args []string) {
	stateJSON, _ := cmd.Flags().GetString("state")
	if stateJSON == "" {
		_ = cmd.Usage()
		log.Fatalf("--state JSON is required")
	}
	var ss core.SignedState
	if err := json.Unmarshal([]byte(stateJSON), &ss); err != nil {
		log.Fatalf("invalid state JSON: %v", err)
	}
	channelBail(engine.InitiateClose(ss))
	fmt.Println("🛑 Close initiated")
}

func challengeHandler(cmd *cobra.Command, args []string) {
	stateJSON, _ := cmd.Flags().GetString("state")
	if stateJSON == "" {
		_ = cmd.Usage()
		log.Fatalf("--state JSON is required")
	}
	var ss core.SignedState
	if err := json.Unmarshal([]byte(stateJSON), &ss); err != nil {
		log.Fatalf("invalid state JSON: %v", err)
	}
	channelBail(engine.Challenge(ss))
	fmt.Println("⚔️  Challenge submitted")
}

func finalizeHandler(cmd *cobra.Command, args []string) {
	cidHex, _ := cmd.Flags().GetString("channel")
	if cidHex == "" {
		_ = cmd.Usage()
		log.Fatalf("--channel id is required")
	}
	idBytes, err := hex.DecodeString(cidHex)
	channelBail(err)
	if len(idBytes) != 32 {
		log.Fatalf("channel id must be 32‑byte hex")
	}
	var id core.ChannelID
	copy(id[:], idBytes)
	channelBail(engine.Finalize(id))
	fmt.Println("✅ Channel finalized")
}

func statusHandler(cmd *cobra.Command, args []string) {
	cidHex, _ := cmd.Flags().GetString("channel")
	if cidHex == "" {
		_ = cmd.Usage()
		log.Fatalf("--channel id is required")
	}
	idBytes, err := hex.DecodeString(cidHex)
	channelBail(err)
	if len(idBytes) != 32 {
		log.Fatalf("channel id must be 32-byte hex")
	}
	var id core.ChannelID
	copy(id[:], idBytes)

	ch, err := engine.GetChannel(id)
	channelBail(err)
	b, _ := json.MarshalIndent(ch, "", "  ")
	fmt.Println(string(b))
}

func listHandler(cmd *cobra.Command, args []string) {
	chans, err := engine.ListChannels()
	channelBail(err)
	b, _ := json.MarshalIndent(chans, "", "  ")
	fmt.Println(string(b))
}

//-------------------------------------------------------------------------
// bail helper
//-------------------------------------------------------------------------

func channelBail(err error) {
	if err != nil {
		log.Fatalf("❌ %v", err)
	}
}

//-------------------------------------------------------------------------
// CLI definitions (top section)
//-------------------------------------------------------------------------

var channelCmd = &cobra.Command{
	Use:              "channel",
	Short:            "Manage off‑chain payment/state channels",
	PersistentPreRun: initMiddleware,
}

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open a new channel between two parties",
	Run:   openChannelHandler,
}

var closeCmd = &cobra.Command{
	Use:   "close",
	Short: "Submit a signed state to start the close process",
	Run:   initiateCloseHandler,
}

var channelChallengeCmd = &cobra.Command{
	Use:   "challenge",
	Short: "Challenge a close with a higher‑nonce state",
	Run:   challengeHandler,
}

var channelFinalizeCmd = &cobra.Command{
	Use:   "finalize",
	Short: "Finalize and settle an expired channel",
	Run:   finalizeHandler,
}

var channelStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current state of a channel",
	Run:   statusHandler,
}

var channelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all open channels",
	Run:   listHandler,
}

//-------------------------------------------------------------------------
// init – flag wiring
//-------------------------------------------------------------------------

func init() {
	// Root persistent flags
	channelCmd.PersistentFlags().String("ledger", "", "Path to ledger database")

	// open flags
	openCmd.Flags().String("partyA", "", "Hex address of party A (20 bytes) [required]")
	openCmd.Flags().String("partyB", "", "Hex address of party B (20 bytes) [required]")
	openCmd.Flags().String("token", "", "Hex token identifier (16 bytes) [required]")
	openCmd.Flags().String("amountA", "0", "Amount deposited by party A [required]")
	openCmd.Flags().String("amountB", "0", "Amount deposited by party B [required]")
	openCmd.Flags().Uint64("nonce", uint64(time.Now().UnixNano()), "Unique nonce (default: current timestamp)")

	// initiate close & challenge flags
	closeCmd.Flags().String("state", "", "Signed state JSON blob [required]")
	channelChallengeCmd.Flags().String("state", "", "Signed state JSON blob [required]")

	// finalize flags
	channelFinalizeCmd.Flags().String("channel", "", "ChannelID in hex (32 bytes) [required]")
	channelStatusCmd.Flags().String("channel", "", "ChannelID in hex (32 bytes) [required]")

	// Sub‑command registration
	channelCmd.AddCommand(openCmd)
	channelCmd.AddCommand(closeCmd)
	channelCmd.AddCommand(channelChallengeCmd)
	channelCmd.AddCommand(channelFinalizeCmd)
	channelCmd.AddCommand(channelStatusCmd)
	channelCmd.AddCommand(channelListCmd)
}

//-------------------------------------------------------------------------
// Consolidated route export (bottom section)
//-------------------------------------------------------------------------

// ChannelRoute is the entry‑point command to be imported by the main CLI.
var ChannelRoute = channelCmd

//-------------------------------------------------------------------------
// END cmd/cli/state_channel.go
//-------------------------------------------------------------------------
