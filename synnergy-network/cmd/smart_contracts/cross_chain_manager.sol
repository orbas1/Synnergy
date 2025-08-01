// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title CrossChainManager demonstrates usage of Synnergy cross-chain opcodes
///        defined in opcode_dispatcher.go.
/// @notice Uses RegisterBridge (0x090001), LockAndMint (0x090004),
///         BurnAndRelease (0x090005) and GetBridge (0x090006)
///         to bridge assets across chains via inline assembly.
contract CrossChainManager {
    /// Register a new bridge configuration on-chain.
    function registerBridge(string memory src, string memory dst, address relayer) external {
        bytes4 opcode = 0x090001; // RegisterBridge
        bytes memory srcBytes = bytes(src);
        bytes memory dstBytes = bytes(dst);
        assembly {
            let total := add(add(0x40, mload(srcBytes)), mload(dstBytes))
            let ptr := mload(0x40)
            mstore(ptr, mload(srcBytes))
            mstore(add(ptr, 0x20), mload(dstBytes))
            mstore(add(ptr, 0x40), relayer)
            calldatacopy(add(ptr, 0x60), add(srcBytes, 32), mload(srcBytes))
            calldatacopy(add(add(ptr, 0x60), mload(srcBytes)), add(dstBytes, 32), mload(dstBytes))
            let success := call(gas(), 0, opcode, ptr, add(total, 0x60), 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }

    /// Lock native assets and mint wrapped tokens on the target chain.
    function lockAndMint(uint32 assetId, uint64 amount, bytes calldata proof) external {
        bytes4 opcode = 0x090004; // LockAndMint
        assembly {
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
    function burnAndRelease(uint32 assetId, address to, uint64 amount) external {
        bytes4 opcode = 0x090005; // BurnAndRelease
        assembly {
            mstore(0x0, assetId)
            mstore(0x20, to)
            mstore(0x40, amount)
            let success := call(gas(), 0, opcode, 0x0, 0x60, 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }

    /// Retrieve a bridge configuration by ID.
    function getBridge(bytes32 id) external returns (bytes memory result) {
        bytes4 opcode = 0x090006; // GetBridge
        assembly {
            mstore(0x0, id)
            let success := staticcall(gas(), 0, opcode, 0x0, 0x20, 0x0, 0x40)
            if iszero(success) { revert(0, 0) }
            result := mload(0x40)
            mstore(result, mload(0x0))
        }
    }
}
