package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core"
)

var (
	btLedgerPath string
	btLedger     *core.Ledger
)

func initBTLedger(cmd *cobra.Command, _ []string) error {
	if btLedger != nil {
		return nil
	}
	path := btLedgerPath
	if path == "" {
		path = viper.GetString("LEDGER_PATH")
	}
	if path == "" {
		return fmt.Errorf("ledger path not specified")
	}
	var err error
	btLedger, err = core.OpenLedger(path)
	if err != nil {
		return fmt.Errorf("open ledger: %w", err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Controllers
// -----------------------------------------------------------------------------

func btHandleCreate(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("create requires tree name")
	}
	_, err := core.BinaryTreeNew(args[0], btLedger)
	return err
}

func btHandleInsert(cmd *cobra.Command, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("insert <tree> <key> <value>")
	}
	val, err := hex.DecodeString(args[2])
	if err != nil {
		val = []byte(args[2])
	}
	return core.BinaryTreeInsert(args[0], args[1], val)
}

func btHandleSearch(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("search <tree> <key>")
	}
	val, ok, err := core.BinaryTreeSearch(args[0], args[1])
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("not found")
		return nil
	}
	fmt.Println(hex.EncodeToString(val))
	return nil
}

func btHandleDelete(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("delete <tree> <key>")
	}
	return core.BinaryTreeDelete(args[0], args[1])
}

func btHandleList(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("list <tree>")
	}
	keys, err := core.BinaryTreeInOrder(args[0])
	if err != nil {
		return err
	}
	for _, k := range keys {
		fmt.Println(k)
	}
	return nil
}

// -----------------------------------------------------------------------------
// CLI definitions
// -----------------------------------------------------------------------------

var btCmd = &cobra.Command{
	Use:               "binarytree",
	Short:             "manage binary trees in the ledger",
	PersistentPreRunE: initBTLedger,
}

var btCreateCmd = &cobra.Command{Use: "create <name>", Short: "create tree", RunE: btHandleCreate}
var btInsertCmd = &cobra.Command{Use: "insert <tree> <key> <value>", Short: "insert value", RunE: btHandleInsert}
var btSearchCmd = &cobra.Command{Use: "search <tree> <key>", Short: "search key", RunE: btHandleSearch}
var btDeleteCmd = &cobra.Command{Use: "delete <tree> <key>", Short: "delete key", RunE: btHandleDelete}
var btListCmd = &cobra.Command{Use: "list <tree>", Short: "list keys", RunE: btHandleList}

func init() {
	btCmd.PersistentFlags().StringVar(&btLedgerPath, "ledger", "", "path to ledger")
	btCmd.AddCommand(btCreateCmd, btInsertCmd, btSearchCmd, btDeleteCmd, btListCmd)
}

// BinaryTreeCmd is the exported root command.
var BinaryTreeCmd = btCmd

func RegisterBinaryTree(root *cobra.Command) { root.AddCommand(BinaryTreeCmd) }
