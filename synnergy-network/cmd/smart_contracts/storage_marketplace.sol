// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title StorageMarketplace interacts with Synnergy's storage module
///        via custom VM opcodes defined in `opcode_dispatcher.go`.
contract StorageMarketplace {
    bytes4 private constant CREATE_LISTING = 0x180004;
    bytes4 private constant OPEN_DEAL      = 0x180006;
    bytes4 private constant CLOSE_DEAL     = 0x180008;
    bytes4 private constant RELEASE_ESCROW = 0x180009;

    /// @notice Create a new storage listing on-chain.
    /// @param listingData ABI-encoded struct expected by the storage module.
    function createListing(bytes memory listingData) external {
        bytes4 opcode = CREATE_LISTING;
        assembly {
            let success := call(gas(), 0, opcode, add(listingData, 32), mload(listingData), 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }

    /// @notice Open a storage deal between a client and provider.
    /// @param dealData ABI-encoded struct describing the deal.
    function openDeal(bytes memory dealData) external {
        bytes4 opcode = OPEN_DEAL;
        assembly {
            let success := call(gas(), 0, opcode, add(dealData, 32), mload(dealData), 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }

    /// @notice Close a deal and release its escrow back to the provider.
    /// @param dealId Identifier of the deal to close.
    /// @param escrowId Associated escrow account ID.
    function closeDeal(bytes32 dealId, bytes32 escrowId) external {
        bytes4 closeOp = CLOSE_DEAL;
        bytes4 releaseOp = RELEASE_ESCROW;
        assembly {
            mstore(0x0, dealId)
            let success := call(gas(), 0, closeOp, 0x0, 0x20, 0, 0)
            if iszero(success) { revert(0, 0) }
            mstore(0x0, escrowId)
            success := call(gas(), 0, releaseOp, 0x0, 0x20, 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }
}
