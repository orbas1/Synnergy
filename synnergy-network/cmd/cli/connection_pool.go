package cli

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"synnergy-network/core"
)

var (
	connPool *core.ConnPool
	cpOnce   sync.Once
)

func cpInit(cmd *cobra.Command, _ []string) error {
	cpOnce.Do(func() {
		d := core.NewDialer(5*time.Second, 30*time.Second)
		connPool = core.NewConnPool(d, 4, time.Minute)
	})
	return nil
}

func cpStats(cmd *cobra.Command, _ []string) error {
	if connPool == nil {
		return fmt.Errorf("connection pool not initialised")
	}
	fmt.Fprintf(cmd.OutOrStdout(), "idle connections: %d\n", connPool.Stats())
	return nil
}

func cpDial(cmd *cobra.Command, args []string) error {
	if connPool == nil {
		return fmt.Errorf("connection pool not initialised")
	}
	if len(args) != 1 {
		return fmt.Errorf("dial requires <addr>")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := connPool.Acquire(ctx, args[0])
	if err != nil {
		return err
	}
	connPool.Release(conn)
	fmt.Fprintln(cmd.OutOrStdout(), "dial ok")
	return nil
}

func cpClose(cmd *cobra.Command, _ []string) error {
	if connPool != nil {
		connPool.Close()
		connPool = nil
	}
	fmt.Fprintln(cmd.OutOrStdout(), "pool closed")
	return nil
}

var connPoolCmd = &cobra.Command{
	Use:               "connpool",
	Short:             "Manage network connection pool",
	PersistentPreRunE: cpInit,
}

func init() {
	connPoolCmd.AddCommand(&cobra.Command{
		Use:   "stats",
		Short: "Show pool statistics",
		RunE:  cpStats,
	})
	connPoolCmd.AddCommand(&cobra.Command{
		Use:   "dial <addr>",
		Short: "Dial address using the pool",
		Args:  cobra.ExactArgs(1),
		RunE:  cpDial,
	})
	connPoolCmd.AddCommand(&cobra.Command{
		Use:   "close",
		Short: "Close the pool",
		RunE:  cpClose,
	})
}

var ConnPoolCmd = connPoolCmd
