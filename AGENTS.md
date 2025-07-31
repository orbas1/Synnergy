# Synnergy Compile Error Checklist

This file tracks compile errors between CLI commands and core modules of the Synnergy network. Use it to coordinate fixes and keep progress organized.

## Setup
1. Run `./setup_synn.sh` to install dependencies and build the CLI.
2. Work from a clean git state before starting fixes.

## Workflow
1. Choose the next unchecked file from the checklist below (work on up to **three files at a time**).
2. Open the file and resolve any compile errors.
3. Run `go vet ./...` and `go build ./...` to verify.
4. Mark the file as completed by checking the box in this document.
5. Commit your changes referencing the files you fixed.

## CLI Files
1. [ ] ai.go
2. [ ] amm.go
3. [ ] authority_node.go
4. [ ] charity_pool.go
5. [ ] coin.go
6. [ ] compliance.go
7. [ ] consensus.go
8. [ ] contracts.go
9. [ ] cross_chain.go
10. [ ] data.go
11. [ ] fault_tolerance.go
12. [ ] governance.go
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
25. [ ] storage.go
26. [ ] tokens.go
27. [ ] transactions.go
28. [ ] utility_functions.go
29. [ ] virtual_machine.go
30. [ ] wallet.go

## Module Files
All modules live under `synnergy-network/core`. For a short description of each
file, see [`synnergy-network/core/module_guide.md`](synnergy-network/core/module_guide.md).
1. [ ] ai.go
2. [ ] amm.go
3. [ ] authority_nodes.go
4. [ ] charity_pool.go
5. [ ] coin.go
6. [ ] common_structs.go
7. [ ] compliance.go
8. [ ] consensus.go
9. [ ] contracts.go
10. [ ] contracts_opcodes.go
11. [ ] cross_chain.go
12. [ ] data.go
13. [ ] fault_tolerance.go
14. [ ] gas_table.go
15. [ ] governance.go
16. [ ] green_technology.go
17. [ ] ledger.go
18. [ ] ledger_test.go
19. [ ] liquidity_pools.go
20. [ ] loanpool.go
21. [ ] network.go
22. [ ] opcode_dispatcher.go
23. [ ] replication.go
24. [ ] rollups.go
25. [ ] security.go
26. [ ] sharding.go
27. [ ] sidechains.go
28. [ ] state_channel.go
29. [ ] storage.go
30. [ ] tokens.go
31. [ ] transactions.go
32. [ ] utility_functions.go
33. [ ] virtual_machine.go
34. [ ] wallet.go

## Guidance
- Always tackle files in numerical order. Start with the lowest numbered unchecked file and address compile errors.
- Work on at most three files before committing your changes.
- After verifying builds succeed, mark the corresponding boxes here.
- Reference `setup_synn.sh` whenever setting up a new environment.
