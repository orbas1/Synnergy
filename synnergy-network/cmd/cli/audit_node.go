package cli

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core"
	"synnergy-network/pkg/utils"
)

var (
	audNode *core.AuditNode
	audMu   sync.RWMutex
)

func auditInit(cmd *cobra.Command, _ []string) error {
	if audNode != nil {
		return nil
	}
	_ = godotenv.Load()
	lv, err := logrus.ParseLevel(viper.GetString("logging.level"))
	if err != nil {
		return err
	}
	logrus.SetLevel(lv)

	netCfg := core.Config{
		ListenAddr:     viper.GetString("network.listen_addr"),
		BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
		DiscoveryTag:   viper.GetString("network.discovery_tag"),
	}
	wal := utils.EnvOrDefault("LEDGER_WAL", "./ledger.wal")
	snap := utils.EnvOrDefault("LEDGER_SNAPSHOT", "./ledger.snap")
	interval := utils.EnvOrDefaultInt("LEDGER_SNAPSHOT_INTERVAL", 100)
	ledCfg := core.LedgerConfig{WALPath: wal, SnapshotPath: snap, SnapshotInterval: interval}
	boot := core.BootstrapConfig{Network: netCfg, Ledger: ledCfg, Replication: nil}
	cfg := core.AuditNodeConfig{Bootstrap: boot, TrailPath: os.Getenv("AUDIT_FILE")}
	n, err := core.NewAuditNode(&cfg)
	if err != nil {
		return err
	}
	audMu.Lock()
	audNode = n
	audMu.Unlock()
	return nil
}

func auditStart(cmd *cobra.Command, _ []string) error {
	audMu.RLock()
	n := audNode
	audMu.RUnlock()
	if n == nil {
		return errors.New("not initialised")
	}
	n.Start()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		_ = n.Stop()
		os.Exit(0)
	}()
	fmt.Fprintln(cmd.OutOrStdout(), "audit node started")
	return nil
}

func auditStop(cmd *cobra.Command, _ []string) error {
	audMu.RLock()
	n := audNode
	audMu.RUnlock()
	if n == nil {
		return errors.New("not running")
	}
	_ = n.Stop()
	audMu.Lock()
	audNode = nil
	audMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func auditLog(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return errors.New("usage: log <addrHex> <event> [meta.json]")
	}
	audMu.RLock()
	n := audNode
	audMu.RUnlock()
	if n == nil {
		return errors.New("not running")
	}
	addrBytes, err := hex.DecodeString(args[0])
	if err != nil || len(addrBytes) != 20 {
		return errors.New("invalid address")
	}
	var addr core.Address
	copy(addr[:], addrBytes)
	meta := map[string]string{}
	if len(args) == 3 {
		raw, err := os.ReadFile(args[2])
		if err != nil {
			return fmt.Errorf("read meta: %w", err)
		}
		if err := json.Unmarshal(raw, &meta); err != nil {
			return fmt.Errorf("decode meta: %w", err)
		}
	}
	if err := n.LogAudit(addr, args[1], meta); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "event recorded")
	return nil
}

func auditEvents(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: events <addrHex>")
	}
	audMu.RLock()
	n := audNode
	audMu.RUnlock()
	if n == nil {
		return errors.New("not running")
	}
	addrBytes, err := hex.DecodeString(args[0])
	if err != nil || len(addrBytes) != 20 {
		return errors.New("invalid address")
	}
	var addr core.Address
	copy(addr[:], addrBytes)
	evs, err := n.AuditEvents(addr)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(evs)
}

func auditArchive(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: archive <dest>")
	}
	audMu.RLock()
	n := audNode
	audMu.RUnlock()
	if n == nil {
		return errors.New("not running")
	}
	path, sum, err := n.ArchiveTrail(args[0])
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "archive written to %s (sha256 %s)\n", path, sum)
	return nil
}

var auditNodeCmd = &cobra.Command{Use: "auditnode", Short: "Run audit node", PersistentPreRunE: auditInit}
var auditNodeStart = &cobra.Command{Use: "start", Short: "Start node", RunE: auditStart}
var auditNodeStop = &cobra.Command{Use: "stop", Short: "Stop node", RunE: auditStop}
var auditNodeLog = &cobra.Command{Use: "log", Short: "Record event", RunE: auditLog}
var auditNodeEvents = &cobra.Command{Use: "events", Short: "List events", RunE: auditEvents}
var auditNodeArchive = &cobra.Command{Use: "archive", Short: "Archive audit trail", RunE: auditArchive}

func init() {
	auditNodeCmd.AddCommand(auditNodeStart, auditNodeStop, auditNodeLog, auditNodeEvents, auditNodeArchive)
}

// AuditNodeCmd exported for index registration
var AuditNodeCmd = auditNodeCmd
