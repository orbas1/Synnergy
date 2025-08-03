package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
	Tokens "synnergy-network/core/Tokens"
)

var (
	etOnce  sync.Once
	etToken *Tokens.EventTicketToken
)

func etInit(cmd *cobra.Command, _ []string) error {
	var err error
	etOnce.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		led, e := core.OpenLedger(path)
		if e != nil {
			err = e
			return
		}
		gas := core.NewFlatGasCalculator(core.DefaultGasPrice)
		meta := core.Metadata{Name: "Event Ticket", Symbol: "SYNTIX", Decimals: 0, Standard: core.StdSYN1700}
		etToken, err = Tokens.NewEventTicketToken(meta, led, gas, nil)
		if err != nil {
			return
		}
	})
	return err
}

func parseAddr(str string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(str)
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func handleCreateEvent(cmd *cobra.Command, _ []string) error {
	name, _ := cmd.Flags().GetString("name")
	desc, _ := cmd.Flags().GetString("desc")
	loc, _ := cmd.Flags().GetString("location")
	start, _ := cmd.Flags().GetInt64("start")
	end, _ := cmd.Flags().GetInt64("end")
	supply, _ := cmd.Flags().GetUint64("supply")
	meta := Tokens.EventMetadata{
		Name:         name,
		Description:  desc,
		Location:     loc,
		StartTime:    time.Unix(start, 0),
		EndTime:      time.Unix(end, 0),
		TicketSupply: supply,
	}
	id := etToken.CreateEvent(meta)
	fmt.Fprintf(cmd.OutOrStdout(), "event created %d\n", id)
	return nil
}

func handleIssue(cmd *cobra.Command, _ []string) error {
	eventID, _ := cmd.Flags().GetUint64("event")
	ownerStr, _ := cmd.Flags().GetString("owner")
	class, _ := cmd.Flags().GetString("class")
	typ, _ := cmd.Flags().GetString("type")
	price, _ := cmd.Flags().GetUint64("price")
	owner, err := parseAddr(ownerStr)
	if err != nil {
		return err
	}
	ticketID, err := etToken.IssueTicket(eventID, owner, Tokens.TicketMetadata{Price: price, Class: class, Type: typ})
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "ticket %d issued\n", ticketID)
	return nil
}

func handleTransferTicket(cmd *cobra.Command, _ []string) error {
	ticketID, _ := cmd.Flags().GetUint64("ticket")
	fromStr, _ := cmd.Flags().GetString("from")
	toStr, _ := cmd.Flags().GetString("to")
	from, err := parseAddr(fromStr)
	if err != nil {
		return err
	}
	to, err := parseAddr(toStr)
	if err != nil {
		return err
	}
	if err := etToken.TransferTicket(ticketID, from, to); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "transfer ok")
	return nil
}

func handleVerify(cmd *cobra.Command, _ []string) error {
	ticketID, _ := cmd.Flags().GetUint64("ticket")
	holderStr, _ := cmd.Flags().GetString("holder")
	holder, err := parseAddr(holderStr)
	if err != nil {
		return err
	}
	ok := etToken.VerifyTicket(ticketID, holder)
	if ok {
		fmt.Fprintln(cmd.OutOrStdout(), "valid")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "invalid")
	}
	return nil
}

var ticketCmd = &cobra.Command{Use: "syn1700", Short: "SYN1700 event tickets", PersistentPreRunE: etInit}
var evtCreateCmd = &cobra.Command{Use: "create-event", Short: "Create event", RunE: handleCreateEvent}
var issueCmd = &cobra.Command{Use: "issue", Short: "Issue ticket", RunE: handleIssue}
var transferTicketCmd = &cobra.Command{Use: "transfer", Short: "Transfer ticket", RunE: handleTransferTicket}
var verifyCmd = &cobra.Command{Use: "verify", Short: "Verify ticket", RunE: handleVerify}

func init() {
	evtCreateCmd.Flags().String("name", "", "name")
	evtCreateCmd.Flags().String("desc", "", "description")
	evtCreateCmd.Flags().String("location", "", "location")
	evtCreateCmd.Flags().Int64("start", time.Now().Unix(), "start time")
	evtCreateCmd.Flags().Int64("end", time.Now().Add(time.Hour).Unix(), "end time")
	evtCreateCmd.Flags().Uint64("supply", 0, "ticket supply")
	issueCmd.Flags().Uint64("event", 0, "event id")
	issueCmd.Flags().String("owner", "", "owner")
	issueCmd.Flags().String("class", "Standard", "class")
	issueCmd.Flags().String("type", "Standard", "type")
	issueCmd.Flags().Uint64("price", 0, "price")
	transferTicketCmd.Flags().Uint64("ticket", 0, "ticket id")
	transferTicketCmd.Flags().String("from", "", "from")
	transferTicketCmd.Flags().String("to", "", "to")
	verifyCmd.Flags().Uint64("ticket", 0, "ticket id")
	verifyCmd.Flags().String("holder", "", "holder")
	ticketCmd.AddCommand(evtCreateCmd, issueCmd, transferTicketCmd, verifyCmd)
}

var TicketCmd = ticketCmd

func RegisterEventTicket(root *cobra.Command) { root.AddCommand(TicketCmd) }
