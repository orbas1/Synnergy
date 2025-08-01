package cli

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	ztLedger string
	ztEngine *core.ZeroTrustEngine
)

func ztMiddleware(cmd *cobra.Command, args []string) {
	_ = godotenv.Load()
	if lp, _ := cmd.Flags().GetString("ledger"); lp != "" {
		ztLedger = lp
	} else if env := os.Getenv("LEDGER_PATH"); env != "" {
		ztLedger = env
	} else {
		exe, _ := os.Executable()
		ztLedger = filepath.Join(filepath.Dir(exe), "state.db")
	}
	led, err := core.NewLedger(core.LedgerConfig{WALPath: ztLedger})
	if err != nil {
		panic(fmt.Errorf("init ledger: %w", err))
	}
	core.InitZeroTrustChannels(led)
	ztEngine = core.ZTChannels()
}

// controller helpers

type ztController struct{}

func (ztc *ztController) Open(a, b string, token string, amtA, amtB, nonce uint64) (core.ZeroTrustChannelID, error) {
	addrA, err := core.ParseAddress(a)
	if err != nil {
		return core.ZeroTrustChannelID{}, err
	}
	addrB, err := core.ParseAddress(b)
	if err != nil {
		return core.ZeroTrustChannelID{}, err
	}
	tok, err := strconv.ParseUint(token, 10, 32)
	if err != nil {
		return core.ZeroTrustChannelID{}, err
	}
	return ztEngine.OpenChannel(addrA, addrB, core.TokenID(tok), amtA, amtB, nonce)
}

func (ztc *ztController) Send(idStr, fromStr, dataHex string) error {
	idBytes, err := hex.DecodeString(idStr)
	if err != nil || len(idBytes) != 32 {
		return errors.New("invalid id")
	}
	var id core.ZeroTrustChannelID
	copy(id[:], idBytes)
	addr, err := core.ParseAddress(fromStr)
	if err != nil {
		return err
	}
	payload, err := hex.DecodeString(dataHex)
	if err != nil {
		return err
	}
	return ztEngine.Send(id, addr, payload)
}

func (ztc *ztController) Close(idStr string) error {
	idBytes, err := hex.DecodeString(idStr)
	if err != nil || len(idBytes) != 32 {
		return errors.New("invalid id")
	}
	var id core.ZeroTrustChannelID
	copy(id[:], idBytes)
	return ztEngine.Close(id)
}

// CLI command declarations

var ztOpenCmd = &cobra.Command{
	Use:   "ztdc open <addrA> <addrB> <token> <amountA> <amountB> <nonce>",
	Short: "Open a zero trust data channel",
	Args:  cobra.ExactArgs(6),
	RunE: func(cmd *cobra.Command, args []string) error {
		amtA, err := strconv.ParseUint(args[3], 10, 64)
		if err != nil {
			return err
		}
		amtB, err := strconv.ParseUint(args[4], 10, 64)
		if err != nil {
			return err
		}
		nonce, err := strconv.ParseUint(args[5], 10, 64)
		if err != nil {
			return err
		}
		ctrl := &ztController{}
		id, err := ctrl.Open(args[0], args[1], args[2], amtA, amtB, nonce)
		if err != nil {
			return err
		}
		fmt.Printf("channel %x opened\n", id)
		return nil
	},
}

var ztSendCmd = &cobra.Command{
	Use:   "ztdc send <channelIDhex> <fromAddr> <payloadHex>",
	Short: "Send encrypted data over a channel",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ztController{}
		return ctrl.Send(args[0], args[1], args[2])
	},
}

var ztCloseCmd = &cobra.Command{
	Use:   "ztdc close <channelIDhex>",
	Short: "Close a zero trust channel and release funds",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ztController{}
		return ctrl.Close(args[0])
	},
}

var ztCmd = &cobra.Command{
	Use:              "ztdc",
	Short:            "Zero trust data channel commands",
	PersistentPreRun: ztMiddleware,
}

func init() {
	ztCmd.PersistentFlags().String("ledger", "", "path to ledger database")
	ztCmd.AddCommand(ztOpenCmd, ztSendCmd, ztCloseCmd)
}

var ZTChannelCmd = ztCmd
