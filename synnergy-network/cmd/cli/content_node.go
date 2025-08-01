package cli

// cmd/cli/content_node.go - CLI for content node operations.

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

func ensureContentInit(cmd *cobra.Command, _ []string) error {
	if core.CurrentStore() == nil {
		return fmt.Errorf("KV store not initialised")
	}
	return nil
}

// Controller

type ContentController struct{}

func (c *ContentController) RegisterNode(addr, host string, capGB int) error {
	a, err := core.ParseAddress(addr)
	if err != nil {
		return err
	}
	n := core.ContentNetworkNode{ID: a, Addr: host, CapacityGB: capGB}
	return core.RegisterContentNode(n)
}

func (c *ContentController) ListNodes() ([]core.ContentNetworkNode, error) {
	return core.ListContentNodes()
}

func (c *ContentController) Upload(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return core.UploadContent(b, []byte("defaultkey123456"))
}

func (c *ContentController) Retrieve(cid, out string) error {
	data, err := core.RetrieveContent(cid)
	if err != nil {
		return err
	}
	if out == "-" {
		os.Stdout.Write(data)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	return ioutil.WriteFile(out, data, 0o644)
}

// Commands

var contentCmd = &cobra.Command{Use: "content", Short: "Content node operations", PersistentPreRunE: ensureContentInit}

var contentRegisterCmd = &cobra.Command{
	Use:   "register <address> <host:port> <capacityGB>",
	Short: "Register content node",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ContentController{}
		capGB, err := strconv.Atoi(args[2])
		if err != nil || capGB <= 0 {
			return fmt.Errorf("capacityGB must be positive int")
		}
		return ctrl.RegisterNode(args[0], args[1], capGB)
	},
}

var contentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List content nodes",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctrl := &ContentController{}
		list, err := ctrl.ListNodes()
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(list, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(enc))
		return nil
	},
}

var contentUploadCmd = &cobra.Command{
	Use:   "upload <filePath>",
	Short: "Upload content file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ContentController{}
		cid, err := ctrl.Upload(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "CID: %s\n", cid)
		return nil
	},
}

var contentRetrieveCmd = &cobra.Command{
	Use:   "retrieve <cid> [output|-]",
	Short: "Retrieve content by CID",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := "-"
		if len(args) == 2 {
			out = args[1]
		}
		ctrl := &ContentController{}
		return ctrl.Retrieve(args[0], out)
	},
}

func init() {
	contentCmd.AddCommand(contentRegisterCmd, contentListCmd, contentUploadCmd, contentRetrieveCmd)
}

var ContentNodeCmd = contentCmd

func RegisterContentNodeCLI(root *cobra.Command) { root.AddCommand(ContentNodeCmd) }
