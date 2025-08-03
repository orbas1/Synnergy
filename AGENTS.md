# AGENTS Instructions

This repository tracks remediation of extensive compile-time errors across the `core` and `cmd/cli` packages of `synnergy-network`.

## Stage Plan for Error Resolution

1. Define missing `Nodes` type in `core/staking_node.go`.
2. Implement `shuffleAddresses` in `core/authority_nodes.go`.
3. Add `RewardHalvingPeriod` constant in `core/coin.go`.
4. Add `InitialReward` constant in `core/coin.go`.
5. Provide `CalculateWeights` method in `core/consensus_adaptive_management.go`.
6. Add `SetWeightConfig` method in `core/consensus_adaptive_management.go`.
7. Implement `getDifficulty` in `core/consensus_difficulty.go`.
8. Implement consensus engine `Start` method in `core/consensus_specific_node.go`.
9. Implement consensus engine `Stop` method in `core/consensus_specific_node.go`.
10. Declare `AddressZero` for DAO staking in `core/dao_staking.go` and related tokens.
11. Re-run `go build -gcflags=all=-e ./core/...` to surface remaining core errors.
12. Resolve further undefined identifiers and missing methods across core modules.
13. Clean up unused or missing imports in core.
14. Run `go vet ./core/...` to catch additional issues.
15. After core builds, run `go build -gcflags=all=-e ./cmd/cli` to reveal CLI errors.
16. Fix undefined symbols and missing implementations within CLI files.
17. Remove unused imports and variables in CLI code.
18. Run `go vet ./cmd/cli` for deeper analysis.
19. Add unit tests for both core and CLI to prevent regressions.
20. Execute integration tests to ensure core and CLI interact correctly.

