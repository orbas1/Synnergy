// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title TokenCreator illustrates how to deploy new tokens on Synnergy
/// @notice Deploys new tokens using custom VM opcodes defined in `opcode_dispatcher.go`.
/// @author Synnergy Team
contract TokenCreator {
    /// @notice Deploy a token with metadata and optional initial supply.
    /// @param standard Token standard identifier.
    /// @param decimals Number of decimals for the token.
    /// @param isFixed Indicates whether the supply is fixed.
    /// @param supply Initial supply to mint.
    function create(
        uint8 standard,
        uint8 decimals,
        bool isFixed,
        uint64 supply
    ) external {
        bytes4 createOp = 0x00190010; // Tokens_Create opcode
        bytes4 mintOp = 0x001C001A;   // MintToken_VM opcode
        // solhint-disable-next-line no-inline-assembly
        assembly {
            // layout: [standard(32)][decimals(32)][isFixed(32)][supply(32)]
            mstore(0x0, standard)
            mstore(0x20, decimals)
            mstore(0x40, isFixed)
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
