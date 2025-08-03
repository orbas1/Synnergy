// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title ExplorerUtils demonstrates reading ledger data via Synnergy opcodes
///        defined in core/opcode_dispatcher.go.
contract ExplorerUtils {
    // Addresses for ledger inspection precompiles
    uint256 private constant LAST_HEIGHT = 0x0E0016; // LastBlockHeight
    uint256 private constant GET_BLOCK   = 0x0E000D; // GetBlock

    /// @notice Return the current block height recorded by the ledger
    function lastBlockHeight() public returns (uint64 height) {
        assembly {
            let out := mload(0x40)
            let success := call(gas(), LAST_HEIGHT, 0, 0, 0, out, 32)
            if iszero(success) { revert(0, 0) }
            height := mload(out)
            mstore(0x40, add(out, 32))
        }
    }

    /// @notice Query a block hash by height
    function blockHash(uint64 h) public returns (bytes32 hash) {
        assembly {
            let ptr := mload(0x40)
            mstore(ptr, h)
            let out := add(ptr, 0x20)
            let success := call(gas(), GET_BLOCK, 0, ptr, 0x20, out, 0x20)
            if iszero(success) { revert(0, 0) }
            hash := mload(out)
            mstore(0x40, add(out, 0x20))
        }
    }
}
