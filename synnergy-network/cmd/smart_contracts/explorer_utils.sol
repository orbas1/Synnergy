// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title ExplorerUtils demonstrates reading ledger data via Synnergy opcodes
///        defined in core/opcode_dispatcher.go.
contract ExplorerUtils {
    // Opcode constants from opcode_dispatcher.go
    bytes3 private constant LAST_HEIGHT = hex"0E0016"; // LastBlockHeight
    bytes3 private constant GET_BLOCK   = hex"0E000D"; // GetBlock

    /// @notice Return the current block height recorded by the ledger
    function lastBlockHeight() public returns (uint64 height) {
        assembly {
            let success := call(gas(), 0, LAST_HEIGHT, 0, 0, 0, 32)
            if iszero(success) { revert(0, 0) }
            height := mload(0)
        }
    }

    /// @notice Query a block hash by height
    function blockHash(uint64 h) public returns (bytes32 hash) {
        assembly {
            mstore(0x0, h)
            let success := call(gas(), 0, GET_BLOCK, 0x0, 0x20, 0x0, 0x20)
            if iszero(success) { revert(0, 0) }
            hash := mload(0x0)
        }
    }
}
