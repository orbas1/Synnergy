// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title CrossChainManager demonstrates usage of Synnergy cross-chain opcodes
///        defined in opcode_dispatcher.go.
/// @notice Uses LockAndMint (0x090004) and BurnAndRelease (0x090005)
///         to bridge assets across chains via inline assembly.
contract CrossChainManager {
    /// Lock native assets and mint wrapped tokens on the target chain.
    /// @param assetId The 32-bit asset identifier of the wrapped asset
    /// @param amount Amount of tokens to mint
    /// @param proof  Serialized SPV proof used by the bridge
    function lockAndMint(uint32 assetId, uint64 amount, bytes calldata proof) external {
        bytes4 opcode = 0x090004; // LockAndMint
        assembly {
            // memory layout: [assetId][amount][proof_offset][proof_len][proof]
            mstore(0x0, assetId)
            mstore(0x20, amount)
            mstore(0x40, 0x60)
            let len := proof.length
            mstore(0x60, len)
            calldatacopy(0x80, proof.offset, len)
            let total := add(len, 0x80)
            let success := call(gas(), 0, opcode, 0x0, total, 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }

    /// Burn wrapped tokens and release the underlying native asset.
    /// @param assetId The 32-bit asset identifier of the wrapped asset
    /// @param to Recipient of the released native asset
    /// @param amount Amount to burn and release
    function burnAndRelease(uint32 assetId, address to, uint64 amount) external {
        bytes4 opcode = 0x090005; // BurnAndRelease
        assembly {
            // layout: [assetId][to][amount]
            mstore(0x0, assetId)
            mstore(0x20, to)
            mstore(0x40, amount)
            let success := call(gas(), 0, opcode, 0x0, 0x60, 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }
}
