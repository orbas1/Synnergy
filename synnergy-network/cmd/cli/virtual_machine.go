package cli

// -----------------------------------------------------------------------------
// vm.go – CLI wrapper for the on‑chain WASM Virtual‑Machine HTTP daemon
// -----------------------------------------------------------------------------
// Public commands (after RegisterVM):
//   vm start        – launch HTTP daemon
//   vm stop         – gracefully shut it down
//   vm status       – show mode / listen / uptime
// -----------------------------------------------------------------------------

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/wasmerio/wasmer-go/wasmer"
	"golang.org/x/time/rate"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"synnergy-network/core"
)

// -----------------------------------------------------------------------------
// Globals – initialised once via vmInit
// -----------------------------------------------------------------------------

var (
	vmState core.StateRW
	vmImpl  core.VM
	vmSrv   *http.Server

	runtimeCtx  context.Context
	runtimeStop context.CancelFunc

	vmMode      string
	vmStartTime time.Time

	vmOnce   sync.Once
	vmLogger = logrus.StandardLogger()
)

// -----------------------------------------------------------------------------
// Lazy middleware
// -----------------------------------------------------------------------------

func vmInit(cmd *cobra.Command, _ []string) error {
	var err error
	vmOnce.Do(func() {
		_ = godotenv.Load()

		// logging
		lvl := vmEnvOr("LOG_LEVEL", "info")
		lv, e := logrus.ParseLevel(lvl)
		if e != nil {
			err = e
			return
		}
		vmLogger.SetLevel(lv)
		vmLogger.SetFormatter(&logrus.JSONFormatter{})

		// env config
		vmMode = vmEnvOr("VM_MODE", "super-light")
		listen := vmEnvOr("VM_LISTEN", ":9090")
		gasLimit := vmEnvOrUint64("VM_GAS", 8_000_000)

		// state backend
		st, e := core.NewInMemory()
		if e != nil {
			err = e
			return
		}
		vmState = st

		// choose implementation
		switch vmMode {
		case "super-light":
			vmImpl = core.NewSuperLightVM(st)
		case "light":
			vmImpl = core.NewLightVM(st, core.NewGasMeter(gasLimit))
		case "heavy":
			vmImpl = core.NewHeavyVM(st, core.NewGasMeter(gasLimit), wasmer.NewEngine())
		default:
			err = fmt.Errorf("invalid VM_MODE %s", vmMode)
			return
		}

		// router
		r := mux.NewRouter()
		r.Use(vmRateLimit)
		r.HandleFunc("/execute", vmExecuteHandler).Methods("POST")

		vmSrv = &http.Server{
			Addr:         listen,
			Handler:      r,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  30 * time.Second,
		}
	})
	return err
}

func vmEnvOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func vmEnvOrUint64(k string, def uint64) uint64 {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

// -----------------------------------------------------------------------------
// HTTP handler & limiter
// -----------------------------------------------------------------------------

var vmLimiter = rate.NewLimiter(200, 100)

func vmRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !vmLimiter.Allow() {
			http.Error(w, "rate limit", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func vmExecuteHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string         `json:"bytecode"`
		Ctx  core.VMContext `json:"ctx"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	code, err := hex.DecodeString(req.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rec, err := vmImpl.Execute(code, &req.Ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rec)
}

// -----------------------------------------------------------------------------
// CLI controllers
// -----------------------------------------------------------------------------

func vmHandleStart(cmd *cobra.Command, _ []string) error {
	if vmSrv == nil {
		return errors.New("middleware not initialised")
	}
	if runtimeCtx != nil {
		fmt.Fprintln(cmd.OutOrStdout(), "vm already running")
		return nil
	}

	runtimeCtx, runtimeStop = context.WithCancel(context.Background())
	go func() {
		if err := vmSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			vmLogger.Fatalf("vm http: %v", err)
		}
	}()
	vmStartTime = time.Now()
	fmt.Fprintf(cmd.OutOrStdout(), "vm started on %s (%s)\n", vmSrv.Addr, vmMode)
	return nil
}

func vmHandleStop(cmd *cobra.Command, _ []string) error {
	if runtimeCtx == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "vm not running")
		return nil
	}
	runtimeStop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = vmSrv.Shutdown(ctx)
	runtimeCtx, runtimeStop = nil, nil
	fmt.Fprintln(cmd.OutOrStdout(), "vm stopped")
	return nil
}

func vmHandleStatus(cmd *cobra.Command, _ []string) error {
	running := runtimeCtx != nil
	uptime := time.Since(vmStartTime).Truncate(time.Second)
	fmt.Fprintf(cmd.OutOrStdout(), "running: %v\nlisten: %s\nmode: %s\nuptime: %s\n", running, vmSrv.Addr, vmMode, uptime)
	return nil
}

// -----------------------------------------------------------------------------
// Cobra command tree
// -----------------------------------------------------------------------------

var vmRootCmd = &cobra.Command{Use: "vm", Short: "On‑chain VM HTTP service", PersistentPreRunE: vmInit}
var vmStartCmd = &cobra.Command{Use: "start", Short: "Start daemon", Args: cobra.NoArgs, RunE: vmHandleStart}
var vmStopCmdVar = &cobra.Command{Use: "stop", Short: "Stop daemon", Args: cobra.NoArgs, RunE: vmHandleStop}
var vmStatusCmd = &cobra.Command{Use: "status", Short: "Status", Args: cobra.NoArgs, RunE: vmHandleStatus}

func init() { vmRootCmd.AddCommand(vmStartCmd, vmStopCmdVar, vmStatusCmd) }

// -----------------------------------------------------------------------------
// Export helper
// -----------------------------------------------------------------------------

var VMCmd = vmRootCmd

func RegisterVM(root *cobra.Command) { root.AddCommand(VMCmd) }
