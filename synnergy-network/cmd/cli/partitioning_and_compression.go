package cli

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

func partHandleCompress(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	out, err := core.CompressData(data)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), base64.StdEncoding.EncodeToString(out))
	return nil
}

func partHandleDecompress(cmd *cobra.Command, args []string) error {
	b, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		return err
	}
	out, err := core.DecompressData(b)
	if err != nil {
		return err
	}
	cmd.OutOrStdout().Write(out)
	return nil
}

func partHandleSplit(cmd *cobra.Command, args []string) error {
	size, _ := cmd.Flags().GetInt("size")
	if size <= 0 {
		return fmt.Errorf("--size must be > 0")
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	parts := core.HorizontalPartition(data, size)
	for i, p := range parts {
		fmt.Fprintf(cmd.OutOrStdout(), "part %d: %s\n", i, base64.StdEncoding.EncodeToString(p))
	}
	return nil
}

var (
	partitionCmd      = &cobra.Command{Use: "partition", Short: "Partitioning and compression utilities"}
	partCompressCmd   = &cobra.Command{Use: "compress <file>", Short: "Compress file", Args: cobra.ExactArgs(1), RunE: partHandleCompress}
	partDecompressCmd = &cobra.Command{Use: "decompress <b64>", Short: "Decompress data", Args: cobra.ExactArgs(1), RunE: partHandleDecompress}
	partSplitCmd      = &cobra.Command{Use: "split <file>", Short: "Split file", Args: cobra.ExactArgs(1), RunE: partHandleSplit}
)

func init() {
	partSplitCmd.Flags().Int("size", 1024, "chunk size in bytes")
	partitionCmd.AddCommand(partCompressCmd, partDecompressCmd, partSplitCmd)
}

var PartitionCmd = partitionCmd

func RegisterPartition(root *cobra.Command) { root.AddCommand(PartitionCmd) }
