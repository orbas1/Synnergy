package cli

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

// DataResourceController exposes high level helpers.
type DataResourceController struct {
	mgr *core.DataResourceManager
}

func newDRController() *DataResourceController {
	return &DataResourceController{mgr: core.NewDataResourceManager()}
}

func drmParseAddr(a string) (core.Address, error) {
	b, err := hex.DecodeString(strings.TrimPrefix(a, "0x"))
	if err != nil || len(b) != 20 {
		return core.Address{}, fmt.Errorf("invalid address")
	}
	var out core.Address
	copy(out[:], b)
	return out, nil
}

func (c *DataResourceController) store(owner, key, file string, gas uint64) error {
	addr, err := drmParseAddr(owner)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return c.mgr.Store(addr, key, data, gas)
}

func (c *DataResourceController) load(owner, key, out string) error {
	addr, err := drmParseAddr(owner)
	if err != nil {
		return err
	}
	data, _, err := c.mgr.Load(addr, key)
	if err != nil {
		return err
	}
	if out == "-" {
		fmt.Printf("%s", string(data))
		return nil
	}
	return ioutil.WriteFile(out, data, 0o644)
}

func (c *DataResourceController) del(owner, key string) error {
	addr, err := drmParseAddr(owner)
	if err != nil {
		return err
	}
	return c.mgr.Delete(addr, key)
}

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Data and resource management utilities",
}

var resStoreCmd = &cobra.Command{
	Use:   "store <owner> <key> <file> <gas>",
	Short: "Store data and set gas limit",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		gas, err := parseUint(args[3])
		if err != nil {
			return err
		}
		return newDRController().store(args[0], args[1], args[2], gas)
	},
}

var resLoadCmd = &cobra.Command{
	Use:   "load <owner> <key> [out|-]",
	Short: "Load data for an owner/key",
	Args:  cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := "-"
		if len(args) == 3 {
			out = args[2]
		}
		return newDRController().load(args[0], args[1], out)
	},
}

var resDelCmd = &cobra.Command{
	Use:   "delete <owner> <key>",
	Short: "Delete stored data",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return newDRController().del(args[0], args[1])
	},
}

func init() {
	resourceCmd.AddCommand(resStoreCmd, resLoadCmd, resDelCmd)
}

// ResourceCmd is exported for registration in the root CLI.
var ResourceCmd = resourceCmd

func parseUint(s string) (uint64, error) {
	var v uint64
	_, err := fmt.Sscan(s, &v)
	return v, err
}
