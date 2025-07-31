# Synnergy Development Playbook

This document outlines the workflow for bringing the Synnergy blockchain to a
fully operational state. It includes environment setup, error checking
methodology and a multi‑stage checklist that divides work across twenty‑five
phases.

## Goals

1. Fix all compile errors across CLI and core modules.
2. Provide unit tests for each package and ensure `go test ./...` passes.
3. Build the `synnergy` binary and launch a local node to confirm end‑to‑end
   functionality.

## Setup

1. Execute `./setup_synn.sh` to install Go and project dependencies.
2. Optionally run `./Synnergy.env.sh` to install extra tools and load variables
   from `.env`.
3. Use `go mod tidy` in `synnergy-network` to ensure module requirements are
   recorded.
4. Always start from a clean git state before applying fixes.

## Workflow

1. Select the next unchecked file from the lists below. Work on **no more than
   three files at a time** to keep diffs small.
2. Check open PRs and this checklist to ensure no other agent is already
   tackling the same stage. If a stage is taken, move to the next unchecked
   group of files.
3. Fix compile and logic issues while keeping commits focused.
4. Format code with `go fmt` on the files you changed.
5. Run `go vet` only on the packages containing your changes (for example
   `go vet ./cmd/...`).
6. Build those same packages with `go build`.
7. Execute `go test` for the packages in your current stage rather than the
   entire repository.
8. If dependencies are missing, run `go mod tidy` then re‑run the checks.
9. When all checks pass, mark the file as completed in this document and commit
   your changes.

## Test File Checks

Unit tests live under `synnergy-network/tests`. After modifying a module,
run `go test` only for that module's package path:

```bash
go test ./synnergy-network/core/<module>
```

This keeps failures limited to your stage.

Use `go vet -tests` for additional static analysis of the tests. Watch for
race conditions, failing assertions and resource leaks. All tests should pass
before moving to the next stage.

## Error Checking

For every file ensure the following issues are addressed:

1. **Syntax errors** – detected by `go build`.
2. **Missing imports or modules** – run `go mod tidy` to update `go.sum`.
3. **Lint warnings** – `go vet` flags suspicious code or misused formats.
4. **Formatting** – run `go fmt` to apply standard Go style.
5. **Unit tests** – `go test` must succeed.

Only proceed when all tools report success.

## Blockchain Usage

Once the project builds, you can start a local node from the
`synnergy-network` directory:

```bash
go build -o synnergy ./cmd/synnergy
./synnergy ledger init --path ./ledger.db
./synnergy network start
```

In another terminal create a wallet and mint tokens:

```bash
./synnergy wallet create --out wallet.json
./synnergy coin mint $(jq -r .address wallet.json) 1000
```

See `cmd/cli/cli_guide.md` for further commands like contract deployment or
transferring tokens.

## 25 Stage Plan

The tasks below break file fixes into twenty‑five stages. Update this list as
you progress.

1. **Stage 1** – CLI: `ai.go`, `amm.go`, `authority_node.go`.
2. **Stage 2** – CLI: `charity_pool.go`, `coin.go`, `compliance.go`.
3. **Stage 3** – CLI: `consensus.go`, `contracts.go`, `cross_chain.go`.
4. **Stage 4** – CLI: `data.go`, `fault_tolerance.go`, `governance.go`.
5. **Stage 5** – CLI: `green_technology.go`, `index.go`, `ledger.go`.
6. **Stage 6** – CLI: `liquidity_pools.go`, `loanpool.go`, `network.go`.
7. **Stage 7** – CLI: `replication.go`, `rollups.go`, `security.go`.
8. **Stage 8** – CLI: `sharding.go`, `sidechain.go`, `state_channel.go`.
9. **Stage 9** – CLI: `storage.go`, `tokens.go`, `transactions.go`.
10. **Stage 10** – CLI: `utility_functions.go`, `virtual_machine.go`, `wallet.go`.
11. **Stage 11** – Module: `ai.go`, `amm.go`, `authority_nodes.go`.
12. **Stage 12** – Module: `charity_pool.go`, `coin.go`, `common_structs.go`.
13. **Stage 13** – Module: `compliance.go`, `consensus.go`, `contracts.go`.
14. **Stage 14** – Module: `contracts_opcodes.go`, `cross_chain.go`, `data.go`. ✅
15. **Stage 15** – Module: `fault_tolerance.go`, `gas_table.go`, `governance.go`.
16. **Stage 16** – Module: `green_technology.go`, `ledger.go`, `ledger_test.go`.
17. **Stage 17** – Module: `liquidity_pools.go`, `loanpool.go`, `network.go`.
18. **Stage 18** – Module: `opcode_dispatcher.go`, `replication.go`, `rollups.go`.
19. **Stage 19** – Module: `security.go`, `sharding.go`, `sidechains.go`. ✅
20. **Stage 20** – Module: `state_channel.go`, `storage.go`, `tokens.go`.
21. **Stage 21** – Module: `transactions.go`, `utility_functions.go`,
    `virtual_machine.go`.
22. **Stage 22** – Module: `wallet.go` and review all preceding fixes.
23. **Stage 23** – Run integration tests across CLI packages.
24. **Stage 24** – Launch a local network and verify node start up.
25. **Stage 25** – Final pass through documentation and ensure all tests pass.

## CLI Files
1. [x] ai.go
2. [x] amm.go
3. [x] authority_node.go
4. [x] charity_pool.go
5. [x] coin.go
6. [x] compliance.go
7. [x] consensus.go
8. [x] contracts.go
9. [x] cross_chain.go
10. [x] data.go
11. [x] fault_tolerance.go
12. [x] governance.go
13. [ ] green_technology.go
14. [ ] index.go
15. [ ] ledger.go
16. [ ] liquidity_pools.go
17. [ ] loanpool.go
18. [ ] network.go
19. [ ] replication.go
20. [ ] rollups.go
21. [ ] security.go
22. [ ] sharding.go
23. [ ] sidechain.go
24. [ ] state_channel.go
25. [x] storage.go
26. [x] tokens.go
27. [x] transactions.go
10. [x] data.go
11. [x] fault_tolerance.go
12. [x] governance.go
13. [x] green_technology.go
14. [x] index.go
15. [x] ledger.go
16. [x] liquidity_pools.go
17. [x] loanpool.go
18. [x] network.go
19. [x] replication.go
20. [x] rollups.go
21. [x] security.go
22. [x] sharding.go
23. [x] sidechain.go
24. [x] state_channel.go
25. [x] storage.go
26. [x] tokens.go
27. [x] transactions.go
28. [x] utility_functions.go
29. [x] virtual_machine.go
30. [x] wallet.go


## Module Files
All modules live under `synnergy-network/core`. For a short description of each
file, see [`synnergy-network/core/module_guide.md`](synnergy-network/core/module_guide.md).
1. [ ] ai.go
2. [ ] amm.go
3. [ ] authority_nodes.go
4. [ ] charity_pool.go
5. [ ] coin.go
6. [ ] common_structs.go
7. [x] compliance.go
8. [x] consensus.go
9. [x] contracts.go
10. [x] contracts_opcodes.go
11. [x] cross_chain.go
12. [x] data.go
13. [x] fault_tolerance.go
14. [x] gas_table.go
15. [x] governance.go
16. [x] green_technology.go
17. [x] ledger.go
18. [x] ledger_test.go
19. [x] liquidity_pools.go
20. [x] loanpool.go
21. [x] network.go
22. [x] opcode_dispatcher.go
23. [x] replication.go
24. [x] rollups.go
25. [x] security.go
26. [x] sharding.go
27. [x] sidechains.go
28. [x] state_channel.go
29. [x] storage.go
30. [x] tokens.go
31. [x] transactions.go
32. [x] utility_functions.go
33. [x] virtual_machine.go
34. [ ] wallet.go

## Guidance
- Always work through the stages in order.
- Only modify up to three files before committing.
- Mark checkboxes once builds and tests succeed.
- Refer to `setup_synn.sh` whenever setting up a new environment.
