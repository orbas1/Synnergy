// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title LedgerInspector showcases additional Synnergy opcodes
///        allowing contracts to query ledger state.
contract LedgerInspector {
    // Addresses of custom Synnergy VM opcodes
    uint256 private constant LAST_HASH = 0x0E0003;  // LastBlockHash
    uint256 private constant BALANCE   = 0x0E0012;  // Ledger_BalanceOf

    /// @notice Returns the hash of the latest block recorded by the ledger
    /// @return hash Latest block hash
    function lastBlockHash() external view returns (bytes32 hash) {
        assembly {
            if iszero(staticcall(gas(), LAST_HASH, 0, 0, 0, 32)) { revert(0, 0) }
            hash := mload(0)
        }
    }

    /// @notice Returns the SYNN token balance of an address
    /// @param addr Address to query
    /// @return bal Token balance
    function balanceOf(address addr) external view returns (uint64 bal) {
        assembly {
            mstore(0, addr)
            if iszero(staticcall(gas(), BALANCE, 0, 20, 0, 32)) { revert(0, 0) }
            bal := mload(0)
        }
    }
}
