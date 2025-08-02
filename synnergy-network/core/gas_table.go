// SPDX-License-Identifier: BUSL-1.1
//
// Synnergy Network – Core ▸ Gas Schedule
// --------------------------------------
// This module exposes the gas pricing table used by the virtual machine.  The
// original project contained a gigantic auto-generated map with numerous
// duplicate entries.  Maintaining that list by hand is error prone and caused
// compilation failures.  The rewritten version derives gas costs from opcode
// categories, ensuring every registered opcode receives a price exactly once.

package core

import "log"

// DefaultGasCost is charged for any opcode that does not have an explicit cost
// assigned.  It is intentionally high to highlight missing price entries during
// development.
const DefaultGasCost uint64 = 100_000

// gasTable maps each Opcode to its base gas cost.  It is populated at start-up
// from the opcode catalogue.
var gasTable map[Opcode]uint64

// categoryGas defines a representative base cost for each opcode category.  The
// values are deliberately coarse – detailed tuning belongs in the generator –
// but they avoid duplicate definitions while keeping the VM functional.
var categoryGas = map[byte]uint64{
	0x01: 5_000,  // AI
	0x02: 4_500,  // AMM
	0x03: 8_000,  // Authority
	0x04: 3_000,  // Charity
	0x05: 2_100,  // Coin
	0x06: 5_000,  // Compliance
	0x07: 2_000,  // Consensus
	0x08: 15_000, // Contracts
	0x09: 20_000, // CrossChain
	0x0A: 10_000, // Data
	// Remaining categories fall back to DefaultGasCost.
}

// initGasTable builds the runtime gas table using the (deduplicated) opcode
// catalogue assembled in opcode_dispatcher.go.
func initGasTable() {
	gasTable = make(map[Opcode]uint64, len(catalogue))
	for _, entry := range catalogue {
		cat := byte(entry.op >> 16)
		cost, ok := categoryGas[cat]
		if !ok {
			cost = DefaultGasCost
		}
		gasTable[entry.op] = cost
	}
}

// GasCost returns the base gas price for the given opcode.  If an opcode was not
// included in the table a default punitive cost is applied and logged once.
func GasCost(op Opcode) uint64 {
	if cost, ok := gasTable[op]; ok {
		return cost
	}
	log.Printf("gas_table: missing cost for opcode %d – charging default", op)
	return DefaultGasCost
}
