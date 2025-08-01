# Synnergy Opcode and Gas Guide

This document explains how opcodes are assigned, priced and executed within the Synnergy virtual machine. It also covers how contracts make use of these opcodes and how gas is consumed efficiently during execution.

## Opcode Overview

Every public function in the core modules is mapped to a unique 24‑bit opcode. The format `0xCCNNNN` splits the code into a **category byte** `CC` and a **numeric index** `NNNN`. Categories correspond to module groups such as `AMM`, `Ledger` or `VirtualMachine`. The complete catalogue is generated automatically and can be inspected in [`opcode_dispatcher.go`](opcode_dispatcher.go). An excerpt of the category map is reproduced below:

```
0x01  AI                     0x0F  Liquidity
0x02  AMM                    0x10  Loanpool
0x03  Authority              0x11  Network
0x04  Charity                0x12  Replication
0x05  Coin                   0x13  Rollups
0x06  Compliance             0x14  Security
0x07  Consensus              0x15  Sharding
0x08  Contracts              0x16  Sidechains
0x09  CrossChain             0x17  StateChannel
0x0A  Data                   0x18  Storage
0x0B  FaultTolerance         0x19  Tokens
0x0C  Governance             0x1A  Transactions
0x0D  GreenTech              0x1B  Utilities
0x0E  Ledger                 0x1C  VirtualMachine
                                0x1D  Wallet
```

Opcodes are resolved at startup and registered with the dispatcher. The helper `ToBytecode` converts a function name into its three‑byte opcode which can then be embedded in a contract or transaction payload. Unknown opcodes are rejected before execution to avoid undefined behaviour.

## Operand Management

Most opcodes operate on a stack‑based calling convention similar to the EVM. Operands are consumed from the VM stack in big‑endian word order. For efficiency:

- **Pop operands once.** Avoid repeatedly reading the same value as this will incur additional gas for stack operations.
- **Keep arguments compact.** Use the smallest integer type that fits and pack multiple values into fixed‑size words where possible.
- **Reuse memory.** When working with byte arrays, allocate once using the VM's memory helpers and modify in place.

These patterns reduce the number of opcode invocations and therefore lower gas consumption.

## Gas Schedule

Gas prices are defined in [`gas_table.go`](gas_table.go). Every opcode has a deterministic base cost reflecting its CPU, storage and network impact. If an opcode is missing from the table the VM falls back to `DefaultGasCost`, which is intentionally punitive. A portion of the table looks like this:

```
SwapExactIn       = 4_500
AddLiquidity      = 5_000
RecordVote        = 3_000
RegisterBridge    = 20_000
NewLedger         = 50_000
opSHA256          = 60
```

Gas is charged **before** the handler executes. Additional dynamic costs (for example per‑word memory fees) are accounted for inside the VM's `GasMeter`. When gas runs out the VM aborts with an `out‑of‑gas` error and all state changes are rolled back.

### Saving Gas

- **Batch operations.** Many modules expose bulk functions which perform work in a single opcode. Group transfers or liquidity changes to avoid repeated calls.
- **Read before write.** Check state using lightweight `Get*` opcodes prior to expensive writes. Failed writes still consume gas.
- **Free refunds.** Certain opcodes such as `SelfDestruct` or storage deletion refund a portion of gas when resources are released. Use them judiciously.

## Opcodes in Contracts

Smart contracts invoke opcodes by embedding the 3‑byte code directly into their bytecode. Examples can be found in `cmd/smart_contracts`. During deployment the bytecode is hashed and stored in the ledger via the `Deploy` opcode. Contracts may include an optional Ricardian manifest which links legal prose to the code hash. The registry stores this JSON document and it can be retrieved with `contracts info <address>`.

Ricardian contracts ensure that on‑chain code corresponds to a specific legal agreement. The manifest includes fields such as `name`, `version`, `author` and `terms`. Tools look up the manifest through the `Ricardian` method defined in [`contracts.go`](contracts.go).

## Opcode Translation

Handlers registered in `opcode_dispatcher.go` are looked up at runtime. The dispatcher performs the following steps:

1. Decode the 3‑byte opcode from the transaction or contract bytecode.
2. Retrieve the bound handler and base gas cost.
3. Deduct gas using `GasCost(op)`.
4. Invoke the handler with an `OpContext` providing access to state and the gas meter.

If an opcode is not recognised the dispatcher returns an error and execution halts.

## Virtual Machine Usage

The [`virtual_machine.go`](virtual_machine.go) file implements several VM flavours (super‑light, light and heavy) that share a common interface. Each VM maintains a `GasMeter` which tracks consumption across opcode calls. State writes and event logs are recorded only if the call completes without exceeding the gas limit.

When a contract or script is executed:

1. The VM loads the bytecode and initialises the execution context.
2. For each instruction it calls `Consume(op)` on the `GasMeter` and dispatches the opcode via `Dispatch`.
3. If gas remains after execution, the unused portion is refunded to the caller.

By following the gas schedule and keeping operands minimal, contracts can be written to run predictably across all nodes.

## Further Reading

- [`WHITEPAPER.md`](../WHITEPAPER.md) – high level overview of the opcode catalogue and gas model.
- [`smart_contract_guide.md`](../smart_contract_guide.md) – writing and deploying contracts on Synnergy.
- [`module_guide.md`](module_guide.md) – descriptions of the core modules referenced by each opcode.

