package cli

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/joho/godotenv"
	logrus "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"synnergy-network/core"
)

var (
	ipfsLog  = logrus.New()
	ipfsCfg  = core.StorageConfig{}
	ipfsInit bool
)

func initIPFSMiddleware(cmd *cobra.Command, _ []string) {
	if ipfsInit {
		return
	}
	_ = godotenv.Load()
	if ipfsCfg.IPFSGateway == "" {
		ipfsCfg.IPFSGateway = os.Getenv("IPFS_GATEWAY")
	}
	if ipfsCfg.CacheDir == "" {
		ipfsCfg.CacheDir = os.TempDir()
	}
	if ipfsCfg.GatewayTimeout == 0 {
		ipfsCfg.GatewayTimeout = 30 * time.Second
	}
	if err := core.InitIPFS(&ipfsCfg, ipfsLog); err != nil {
		panic(err)
	}
	ipfsInit = true
}

func addFileHandler(cmd *cobra.Command, args []string) {
	file, _ := cmd.Flags().GetString("file")
	payerHex, _ := cmd.Flags().GetString("payer")
	if file == "" || payerHex == "" {
		_ = cmd.Usage()
		panic("--file and --payer required")
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	addr, err := core.ParseAddress(payerHex)
	if err != nil {
		panic(err)
	}
	cid, err := core.AddFile(context.Background(), data, addr)
	if err != nil {
		panic(err)
	}
	fmt.Println(cid)
}

func getFileHandler(cmd *cobra.Command, args []string) {
	cid, _ := cmd.Flags().GetString("cid")
	out, _ := cmd.Flags().GetString("out")
	if cid == "" {
		_ = cmd.Usage()
		panic("--cid required")
	}
	data, err := core.GetFile(context.Background(), cid)
	if err != nil {
		panic(err)
	}
	if out == "-" || out == "" {
		os.Stdout.Write(data)
		return
	}
	if err := os.WriteFile(out, data, 0o644); err != nil {
		panic(err)
	}
}

func unpinHandler(cmd *cobra.Command, args []string) {
	cid, _ := cmd.Flags().GetString("cid")
	if cid == "" {
		_ = cmd.Usage()
		panic("--cid required")
	}
	if err := core.UnpinFile(context.Background(), cid); err != nil {
		panic(err)
	}
	fmt.Println("unpinned")
}

var ipfsCmd = &cobra.Command{
	Use:              "ipfs",
	Short:            "Interact with the IPFS gateway",
	PersistentPreRun: initIPFSMiddleware,
}

var ipfsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a file to IPFS",
	Run:   addFileHandler,
}

var ipfsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve a file by CID",
	Run:   getFileHandler,
}

var ipfsUnpinCmd = &cobra.Command{
	Use:   "unpin",
	Short: "Unpin a CID from the gateway",
	Run:   unpinHandler,
}

func init() {
	ipfsAddCmd.Flags().String("file", "", "File to upload [required]")
	ipfsAddCmd.Flags().String("payer", "", "Address paying storage rent [required]")
	ipfsGetCmd.Flags().String("cid", "", "CID to fetch [required]")
	ipfsGetCmd.Flags().String("out", "-", "Output file or '-' for STDOUT")
	ipfsUnpinCmd.Flags().String("cid", "", "CID to unpin [required]")

	ipfsCmd.AddCommand(ipfsAddCmd)
	ipfsCmd.AddCommand(ipfsGetCmd)
	ipfsCmd.AddCommand(ipfsUnpinCmd)
}

// IPFSRoute is the exported command group.
var IPFSRoute = ipfsCmd
