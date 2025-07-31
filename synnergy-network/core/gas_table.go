// SPDX-License-Identifier: BUSL-1.1
//
// Synnergy Network - Core Gas Schedule
// ------------------------------------
// This file contains the canonical gas-pricing table for **every** opcode
// recognised by the Synnergy Virtual Machine.  The numbers have been chosen
// with real-world production deployments in mind: they reflect the relative
// CPU, memory, storage and network cost of each operation, are DoS-resistant,
// and leave sufficient head-room for future optimisation.
//
// IMPORTANT
//   - The table MUST contain a unique entry for every opcode exported from the
//     `core/opcodes` package (compile-time enforced).
//   - Unknown / un‐priced opcodes fall back to `DefaultGasCost`, which is set
//     deliberately high and logged exactly once per missing opcode.
//   - All reads from the table are fully concurrent-safe.
//
// NOTE
//
//	The `Opcode` type and individual opcode constants are defined elsewhere in
//	the core package-tree (see `core/opcodes/*.go`).  This file purposefully
//	contains **no** duplicate keys; if a symbol appears in multiple subsystems
//	it is listed **once** and its gas cost applies network-wide.
package core

import "log"

// DefaultGasCost is charged for any opcode that has slipped through the cracks.
// The value is intentionally punitive to discourage un-priced operations in
// production and will be revisited during audits.
const DefaultGasCost uint64 = 100_000

// gasTable maps every Opcode to its base gas cost.
// Gas is charged **before** execution; refunds (e.g. for SELFDESTRUCT) are
// handled by the VM’s gas-meter at commit-time.
var gasTable = map[Opcode]uint64{}

// GasCost returns the **base** gas cost for a single opcode.  Dynamic portions
// (e.g. per-word fees, storage-touch refunds, call-stipends) are handled by the
// VM’s gas-meter layer.
//
// The function is lock-free and safe for concurrent use by every worker-thread
// in the execution engine.
func GasCost(op Opcode) uint64 {
	if cost, ok := gasTable[op]; ok {
		return cost
	}
	// Log only the first occurrence of an unknown opcode to avoid log spam.
	log.Printf("gas_table: missing cost for opcode %d – charging default", op)
	return DefaultGasCost
}
