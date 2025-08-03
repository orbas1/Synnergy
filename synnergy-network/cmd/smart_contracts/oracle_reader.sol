// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title OracleReader fetches data from Synnergy oracles via opcode 0x0A0008
/// @notice Gas cost for QueryOracle is 3,000 as defined in `gas_table.go`.
contract OracleReader {
    /// Query an oracle ID and return raw bytes from the VM.
    function query(bytes32 oracleId) external view returns (bytes memory result) {
        bytes4 opcode = 0x000A0008; // QueryOracle
        assembly {
            mstore(0x0, oracleId)
            // Call the oracle precompile and copy the return data
            if iszero(staticcall(gas(), opcode, 0x0, 0x20, 0x0, 0x0)) { revert(0,0) }
            let size := returndatasize()
            result := mload(0x40)
            mstore(0x40, add(result, add(size, 0x20)))
            mstore(result, size)
            returndatacopy(add(result, 0x20), 0x0, size)
        }
    }
}
