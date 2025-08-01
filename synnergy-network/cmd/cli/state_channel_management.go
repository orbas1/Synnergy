package cli

import (
	"encoding/hex"
	"encoding/json"
	"log"

	"github.com/spf13/cobra"
	"synnergy-network/core"
)

// utilise middleware from state_channel.go

func pauseHandler(cmd *cobra.Command, args []string) {
	cidHex, _ := cmd.Flags().GetString("channel")
	idBytes, err := hex.DecodeString(cidHex)
	if err != nil || len(idBytes) != 32 {
		log.Fatalf("invalid channel id")
	}
	var id core.ChannelID
	copy(id[:], idBytes)
	channelBail(core.Channels().PauseChannel(id))
	log.Println("⏸️  channel paused")
}

func resumeHandler(cmd *cobra.Command, args []string) {
	cidHex, _ := cmd.Flags().GetString("channel")
	idBytes, err := hex.DecodeString(cidHex)
	if err != nil || len(idBytes) != 32 {
		log.Fatalf("invalid channel id")
	}
	var id core.ChannelID
	copy(id[:], idBytes)
	channelBail(core.Channels().ResumeChannel(id))
	log.Println("▶️  channel resumed")
}

func cancelHandler(cmd *cobra.Command, args []string) {
	cidHex, _ := cmd.Flags().GetString("channel")
	idBytes, err := hex.DecodeString(cidHex)
	if err != nil || len(idBytes) != 32 {
		log.Fatalf("invalid channel id")
	}
	var id core.ChannelID
	copy(id[:], idBytes)
	channelBail(core.Channels().CancelClose(id))
	log.Println("🚫 close cancelled")
}

func forceCloseHandler(cmd *cobra.Command, args []string) {
	stateJSON, _ := cmd.Flags().GetString("state")
	if stateJSON == "" {
		_ = cmd.Usage()
		log.Fatalf("--state required")
	}
	var ss core.SignedState
	if err := json.Unmarshal([]byte(stateJSON), &ss); err != nil {
		log.Fatalf("invalid state JSON: %v", err)
	}
	channelBail(core.Channels().ForceClose(ss))
	log.Println("✅ channel forcibly closed")
}

var chanMgmtCmd = &cobra.Command{
	Use:              "channel-mgmt",
	Short:            "Advanced state channel management",
	PersistentPreRun: channelMiddleware,
}

var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause updates on a channel",
	Run:   pauseHandler,
}

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume a paused channel",
	Run:   resumeHandler,
}

var cancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel a pending close",
	Run:   cancelHandler,
}

var forceCloseCmd = &cobra.Command{
	Use:   "force-close",
	Short: "Immediately settle a channel",
	Run:   forceCloseHandler,
}

func init() {
	pauseCmd.Flags().String("channel", "", "ChannelID in hex [required]")
	resumeCmd.Flags().String("channel", "", "ChannelID in hex [required]")
	cancelCmd.Flags().String("channel", "", "ChannelID in hex [required]")
	forceCloseCmd.Flags().String("state", "", "Signed state JSON [required]")

	chanMgmtCmd.AddCommand(pauseCmd)
	chanMgmtCmd.AddCommand(resumeCmd)
	chanMgmtCmd.AddCommand(cancelCmd)
	chanMgmtCmd.AddCommand(forceCloseCmd)
}

// ChannelMgmtRoute exported for registration in index.go
var ChannelMgmtRoute = chanMgmtCmd
