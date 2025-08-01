package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"synnergy-network/core"
)

var msgQueue = core.NewMessageQueue()

func msgEnqueue(cmd *cobra.Command, args []string) error {
	if len(args) != 5 {
		return fmt.Errorf("usage: enqueue <src> <dst> <topic> <type> <payload>")
	}
	src, err := core.ParseAddress(args[0])
	if err != nil {
		return err
	}
	dst, err := core.ParseAddress(args[1])
	if err != nil {
		return err
	}
	payload, err := core.ParseHexPayload(args[4])
	if err != nil {
		return err
	}
	msg := core.NetworkMessage{
		Source:    src,
		Target:    dst,
		Topic:     args[2],
		MsgType:   args[3],
		Content:   payload,
		Timestamp: time.Now().Unix(),
	}
	msgQueue.Enqueue(msg)
	fmt.Fprintln(cmd.OutOrStdout(), "queued")
	return nil
}

func msgProcess(cmd *cobra.Command, _ []string) error {
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	return msgQueue.ProcessNext(led, nil, nil)
}

func msgBroadcast(cmd *cobra.Command, _ []string) error {
	return msgQueue.BroadcastNext()
}

var msgRootCmd = &cobra.Command{Use: "messages", Short: "Message queue"}
var msgEnqCmd = &cobra.Command{Use: "enqueue <src> <dst> <topic> <type> <payload>", Short: "Queue a message", RunE: msgEnqueue}
var msgProcCmd = &cobra.Command{Use: "process", Short: "Process next", RunE: msgProcess}
var msgBroadCmd = &cobra.Command{Use: "broadcast", Short: "Broadcast next", RunE: msgBroadcast}

func init() { msgRootCmd.AddCommand(msgEnqCmd, msgProcCmd, msgBroadCmd) }

// MessagesCmd exposes the command tree for registration in the root CLI.
var MessagesCmd = msgRootCmd

func RegisterMessages(root *cobra.Command) { root.AddCommand(MessagesCmd) }
