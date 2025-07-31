// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title OracleReader fetches data from Synnergy oracles via opcode 0x0A0008
/// @notice Gas cost for QueryOracle is 3,000 as defined in `gas_table.go`.
contract OracleReader {
    /// Query an oracle ID and return raw bytes from the VM.
    function query(bytes32 oracleId) external returns (bytes memory result) {
        bytes4 opcode = 0x0A0008; // QueryOracle
        assembly {
            mstore(0x0, oracleId)
            // We allocate 1K of return buffer for example purposes
            let success := call(gas(), 0, opcode, 0x0, 0x20, 0x100, 0x400)
            if iszero(success) { revert(0, 0) }
            result := mload(0x100)
        }
    }
}
