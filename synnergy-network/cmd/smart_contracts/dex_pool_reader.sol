// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title DexPoolReader retrieves all AMM pools using custom opcodes
/// @notice Utilises opcode 0x0F0008 (Liquidity_Pools) from opcode_dispatcher.go
contract DexPoolReader {
    function pools() external returns (bytes memory data) {
        bytes4 opcode = 0x0F0008;
        assembly {
            let success := call(gas(), 0, opcode, 0, 0, 0, 0)
            if iszero(success) { revert(0, 0) }
            let size := returndatasize()
            data := mload(0x40)
            mstore(data, size)
            let ptr := add(data, 0x20)
            returndatacopy(ptr, 0, size)
            mstore(0x40, add(ptr, size))
        }
    }
}
