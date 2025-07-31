package cli

// ──────────────────────────────────────────────────────────────────────────────
// Synthron Consensus CLI – full‑stack edition (no stubs)
//
// This command set provides production‑grade control over the hybrid PoH/PoS/PoW
// consensus engine.  All placeholder adapters have been replaced with concrete
// services from the synnergy‑network stack (network, security, txpool, authority).
//
// Environment requirements (add these to your .env or orchestration layer):
//   • LEDGER_PATH           – path to Bolt/Badger ledger DB (shared with coin CLI).
//   • KEYSTORE_PATH         – directory with node.key / validator.key PEM (sec svc).
//   • P2P_PORT              – TCP port to listen for p2p (default 30333 if unset).
//   • P2P_BOOTNODES         – comma‑separated multiaddr list of bootnodes.
//   • AUTH_DB_PATH          – path to authority/validator DB (SQLite, Bolt, etc.).
//   • LOG_LEVEL             – trace|debug|info|warn|error (default info).
//   • CONSENSUS_AUTO_START  – "true" to auto‑start engine when CLI initialises.
//
// Wiring into root CLI remains identical:
//     import "synnergy-network/cmd/cli/middleware" // adjust path
//     func init() { middleware.RegisterConsensus(rootCmd) }
//
// ──────────────────────────────────────────────────────────────────────────────

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	// Core & service layer
	"synnergy-network/core"
)

// ──────────────────────────────────────────────────────────────────────────────
// Globals & middleware
// ──────────────────────────────────────────────────────────────────────────────

var (
	consensus   *core.SynnergyConsensus
	consensusMu sync.RWMutex // guards consensus pointer & ctx
	ctx         context.Context
	cancelFn    context.CancelFunc
	ledger      *core.Ledger
	logger      = logrus.StandardLogger()
)

// initConsensusMiddleware loads configuration, initialises shared services
// (ledger, p2p, security, txpool, authority) and constructs the consensus
// engine. It is executed once per process via PersistentPreRunE.
func initConsensusMiddleware(cmd *cobra.Command, _ []string) error {
	var err error

	// 1. env (.env is optional – orchestration layer may set vars directly)
	_ = godotenv.Load()

	// 2. logging level
	lvlStr := os.Getenv("LOG_LEVEL")
	if lvlStr == "" {
		lvlStr = "info"
	}
	lvl, err := logrus.ParseLevel(lvlStr)
	if err != nil {
		return fmt.Errorf("invalid LOG_LEVEL %s: %w", lvlStr, err)
	}
	logger.SetLevel(lvl)

	// 3. ledger (singleton)
	lp := os.Getenv("LEDGER_PATH")
	if lp == "" {
		return fmt.Errorf("LEDGER_PATH not set")
	}
	if ledger == nil {
		ledger, err = core.OpenLedger(lp)
		if err != nil {
			return fmt.Errorf("open ledger: %w", err)
		}
	}

	// Short‑circuit if consensus already initialised
	consensusMu.RLock()
	if consensus != nil {
		consensusMu.RUnlock()
		return nil
	}
	consensusMu.RUnlock()

	// 4. p2p network service
	listenPort := os.Getenv("P2P_PORT")
	if listenPort == "" {
		listenPort = "30333"
	}
	boot := strings.Split(os.Getenv("P2P_BOOTNODES"), ",")
	p2pCfg := network.Config{
		ListenAddr: fmt.Sprintf(":%s", listenPort),
		Bootnodes:  boot,
		Logger:     logger,
	}
	p2pSvc, err := core.NewService(p2pCfg)
	if err != nil {
		return fmt.Errorf("init p2p: %w", err)
	}

	// 5. security (crypto keys & signatures)
	ksPath := os.Getenv("KEYSTORE_PATH")
	if ksPath == "" {
		return fmt.Errorf("KEYSTORE_PATH not set")
	}
	secSvc, err := security.NewService(security.Config{KeyStorePath: ksPath})
	if err != nil {
		return fmt.Errorf("init security: %w", err)
	}

	// 6. transaction pool (bounded‑size mempool driven by ledger state)
	txPoolCfg := txpool.Config{MaxPool: 50_000, Ledger: ledger, Logger: logger}
	txPoolSvc := txpool.New(txPoolCfg)

	// 7. authority (staking / roles)
	authDB := os.Getenv("AUTH_DB_PATH")
	if authDB == "" {
		return fmt.Errorf("AUTH_DB_PATH not set")
	}
	authSvc, err := authority.New(authority.Config{DBPath: authDB, Ledger: ledger})
	if err != nil {
		return fmt.Errorf("init authority: %w", err)
	}

	// 8. create consensus
	cns, err := core.NewConsensus(logger, ledger, p2pSvc, secSvc, txPoolSvc, authSvc)
	if err != nil {
		return fmt.Errorf("new consensus: %w", err)
	}

	consensusMu.Lock()
	consensus = cns
	consensusMu.Unlock()

	// 9. optional auto‑start
	if os.Getenv("CONSENSUS_AUTO_START") == "true" {
		return startConsensus(cmd, nil)
	}

	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Controller helpers
// ──────────────────────────────────────────────────────────────────────────────

func startConsensus(cmd *cobra.Command, _ []string) error {
	consensusMu.Lock()
	defer consensusMu.Unlock()

	if consensus == nil {
		return fmt.Errorf("consensus not initialised – invoke any sub‑command once to bootstrap")
	}
	if ctx != nil {
		fmt.Fprintln(cmd.OutOrStdout(), "consensus already running")
		return nil
	}

	ctx, cancelFn = context.WithCancel(context.Background())
	consensus.Start(ctx)

	// Handle SIGINT/SIGTERM so standalone CLI sessions exit gracefully.
	go func() {
		sigC := make(chan os.Signal, 1)
		signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
		<-sigC
		_ = stopConsensus(cmd, nil)
		os.Exit(0)
	}()

	fmt.Fprintln(cmd.OutOrStdout(), "consensus started ✔")
	return nil
}

func stopConsensus(cmd *cobra.Command, _ []string) error {
	consensusMu.Lock()
	defer consensusMu.Unlock()

	if ctx == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "consensus not running")
		return nil
	}
	cancelFn()
	ctx, cancelFn = nil, nil
	fmt.Fprintln(cmd.OutOrStdout(), "consensus stopped ✔")
	return nil
}

func infoConsensus(cmd *cobra.Command, _ []string) error {
	consensusMu.RLock()
	running := ctx != nil
	consensusMu.RUnlock()

	lastBlk := ledger.LastBlockHeight()
	lastSub := ledger.LastSubBlockHeight()

	fmt.Fprintf(cmd.OutOrStdout(), "running: %v\nsub‑block height: %d\nblock height:    %d\n", running, lastSub, lastBlk)
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Cobra commands (primary definitions first)
// ──────────────────────────────────────────────────────────────────────────────

var consensusCmd = &cobra.Command{
	Use:               "consensus",
	Short:             "Control the hybrid PoH/PoS/PoW consensus engine",
	PersistentPreRunE: initConsensusMiddleware,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Launch consensus loops (non‑blocking)",
	Args:  cobra.NoArgs,
	RunE:  startConsensus,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Gracefully stop consensus loops",
	Args:  cobra.NoArgs,
	RunE:  stopConsensus,
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show consensus height & running status",
	Args:  cobra.NoArgs,
	RunE:  infoConsensus,
}

var weightsCmd = &cobra.Command{
	Use:   "weights [demand] [stake]",
	Short: "Calculate dynamic consensus weights",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		demand, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid demand: %w", err)
		}
		stake, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid stake: %w", err)
		}
		consensusMu.RLock()
		c := consensus
		consensusMu.RUnlock()
		if c == nil {
			return fmt.Errorf("consensus not initialised")
		}
		w := c.CalculateWeights(demand, stake)
		fmt.Fprintf(cmd.OutOrStdout(), "PoW: %.2f%%\nPoS: %.2f%%\nPoH: %.2f%%\n",
			w.PoW*100, w.PoS*100, w.PoH*100)
		return nil
	},
}

func init() {
	consensusCmd.AddCommand(startCmd)
	consensusCmd.AddCommand(stopCmd)
	consensusCmd.AddCommand(infoCmd)
	consensusCmd.AddCommand(weightsCmd)
}

// ──────────────────────────────────────────────────────────────────────────────
// Consolidated route (export)
// ──────────────────────────────────────────────────────────────────────────────

// ConsensusCmd exposes the root command; mirrors naming of other modules.
var ConsensusCmd = consensusCmd

// RegisterConsensus attaches the consensus CLI to any supplied root command.
func RegisterConsensus(root *cobra.Command) { root.AddCommand(ConsensusCmd) }
