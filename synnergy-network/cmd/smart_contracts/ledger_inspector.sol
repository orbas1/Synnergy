// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title LedgerInspector showcases additional Synnergy opcodes
///        allowing contracts to query ledger state.
contract LedgerInspector {
    bytes3 private constant LAST_HASH = hex"0E0003";  // LastBlockHash
    bytes3 private constant BALANCE   = hex"0E0012";  // Ledger_BalanceOf

    /// Returns the hash of the latest block recorded by the ledger
    function lastBlockHash() public returns (bytes32 hash) {
        assembly {
            let success := call(gas(), 0, LAST_HASH, 0, 0, 0, 32)
            if iszero(success) { revert(0, 0) }
            hash := mload(0)
        }
    }

    /// Returns the SYNN token balance of an address
    function balanceOf(bytes20 addr) public returns (uint64 bal) {
        assembly {
            mstore(0, addr)
            let success := call(gas(), 0, BALANCE, 0, 20, 0, 32)
            if iszero(success) { revert(0, 0) }
            bal := mload(0)
        }
    }
}
