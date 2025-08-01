// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title TokenCreator illustrates how to deploy new tokens on Synnergy
///        using custom VM opcodes defined in `opcode_dispatcher.go`.
contract TokenCreator {
    /// Deploy a token with metadata and optional initial supply.
    /// Parameters loosely map to core.Metadata in Go.
    function create(
        uint8 standard,
        uint8 decimals,
        bool fixed,
        uint64 supply
    ) external {
        bytes4 createOp = 0x190010; // Tokens_Create
        bytes4 mintOp = 0x1C001A;   // MintToken_VM
        assembly {
            // layout: [standard(32)][decimals(32)][fixed(32)][supply(32)]
            mstore(0x0, standard)
            mstore(0x20, decimals)
            mstore(0x40, fixed)
            mstore(0x60, supply)
            let success := call(gas(), 0, createOp, 0x0, 0x80, 0, 0)
            if iszero(success) { revert(0, 0) }
            if gt(supply, 0) {
                // assumes tokenId returned in result slot 0x0 (simplified)
                mstore(0x20, caller())
                mstore(0x40, supply)
                success := call(gas(), 0, mintOp, 0x0, 0x60, 0, 0)
                if iszero(success) { revert(0, 0) }
            }
        }
    }
}
