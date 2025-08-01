package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var resourceCmd = &cobra.Command{
	Use:   "resources",
	Short: "Manage VM resource quotas",
}

var rmSetCmd = &cobra.Command{Use: "set <addr> <cpu> <mem> <store>", Short: "Set quota", Args: cobra.ExactArgs(4), RunE: rmHandleSet}
var rmShowCmd = &cobra.Command{Use: "show <addr>", Short: "Show quota", Args: cobra.ExactArgs(1), RunE: rmHandleShow}
var rmChargeCmd = &cobra.Command{Use: "charge <addr> <cpu> <mem> <store>", Short: "Charge usage", Args: cobra.ExactArgs(4), RunE: rmHandleCharge}

func init() {
	resourceCmd.AddCommand(rmSetCmd, rmShowCmd, rmChargeCmd)
}

var ResourceCmd = resourceCmd

func rmParseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func rmHandleSet(_ *cobra.Command, args []string) error {
	addr, err := rmParseAddr(args[0])
	if err != nil {
		return err
	}
	cpu, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	mem, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return err
	}
	sto, err := strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		return err
	}
	return core.RM().SetQuota(addr, cpu, mem, sto)
}

func rmHandleShow(_ *cobra.Command, args []string) error {
	addr, err := rmParseAddr(args[0])
	if err != nil {
		return err
	}
	q, err := core.RM().GetQuota(addr)
	if err != nil {
		return err
	}
	fmt.Printf("CPU:%d MEM:%d STORE:%d USED_CPU:%d USED_MEM:%d USED_STORE:%d\n", q.CPU, q.Memory, q.Storage, q.UsedCPU, q.UsedMem, q.UsedSto)
	return nil
}

func rmHandleCharge(_ *cobra.Command, args []string) error {
	addr, err := rmParseAddr(args[0])
	if err != nil {
		return err
	}
	cpu, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	mem, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return err
	}
	sto, err := strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		return err
	}
	return core.RM().ChargeResources(addr, cpu, mem, sto)
}
