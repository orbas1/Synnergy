package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	eventsCmd     = &cobra.Command{Use: "events", Short: "Manage on-chain events", PersistentPreRunE: eventsInit}
	eventsEmitCmd = &cobra.Command{Use: "emit <type> <data>", Short: "Emit custom event", Args: cobra.ExactArgs(2), RunE: eventsEmit}
	eventsListCmd = &cobra.Command{Use: "list <type>", Short: "List events", Args: cobra.ExactArgs(1), RunE: eventsList}
	eventsGetCmd  = &cobra.Command{Use: "get <type> <id>", Short: "Get event by ID", Args: cobra.ExactArgs(2), RunE: eventsGet}
)

func eventsInit(cmd *cobra.Command, _ []string) error {
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	core.InitEvents(led)
	return nil
}

func eventsEmit(cmd *cobra.Command, args []string) error {
	typ := args[0]
	data := []byte(args[1])
	id, err := core.Events().Emit(&core.Context{}, typ, data)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), id)
	return nil
}

func eventsList(cmd *cobra.Command, args []string) error {
	typ := args[0]
	limit, _ := cmd.Flags().GetInt("limit")
	evs, err := core.Events().List(typ, limit)
	if err != nil {
		return err
	}
	out, _ := json.MarshalIndent(evs, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(out))
	return nil
}

func eventsGet(cmd *cobra.Command, args []string) error {
	typ, id := args[0], args[1]
	ev, err := core.Events().Get(typ, id)
	if err != nil {
		return err
	}
	out, _ := json.MarshalIndent(ev, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(out))
	return nil
}

func init() {
	eventsListCmd.Flags().Int("limit", 0, "max items")
	eventsCmd.AddCommand(eventsEmitCmd, eventsListCmd, eventsGetCmd)
}

var EventsCmd = eventsCmd

func RegisterEvents(root *cobra.Command) { root.AddCommand(EventsCmd) }
