// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title StorageMarketplace interacts with Synnergy's storage module
///        via custom VM opcodes defined in `opcode_dispatcher.go`.
contract StorageMarketplace {
    bytes4 private constant NEW_STORAGE    = 0x180001;
    bytes4 private constant STORAGE_PIN    = 0x180002;
    bytes4 private constant STORAGE_RETRV  = 0x180003;
    bytes4 private constant CREATE_LISTING = 0x180004;
    bytes4 private constant EXISTS         = 0x180005;
    bytes4 private constant OPEN_DEAL      = 0x180006;
    bytes4 private constant STORAGE_CREATE = 0x180007;
    bytes4 private constant CLOSE_DEAL     = 0x180008;
    bytes4 private constant RELEASE_ESCROW = 0x180009;
    bytes4 private constant GET_LISTING    = 0x18000A;
    bytes4 private constant LIST_LISTINGS  = 0x18000B;
    bytes4 private constant GET_DEAL       = 0x18000C;
    bytes4 private constant LIST_DEALS     = 0x18000D;


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

    /// @notice Initialize a new storage space for a provider.
    function newStorage(bytes memory cfg) external {
        bytes4 opcode = NEW_STORAGE;
        assembly {
            let success := call(gas(), 0, opcode, add(cfg, 32), mload(cfg), 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }

    /// @notice Pin a CID into the storage network.
    function pin(bytes memory cid) external {
        bytes4 opcode = STORAGE_PIN;
        assembly {
            let success := call(gas(), 0, opcode, add(cid, 32), mload(cid), 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }

    /// @notice Retrieve content from storage by CID.
    function retrieve(bytes memory cid) external returns (bytes memory result) {
        bytes4 opcode = STORAGE_RETRV;
        assembly {
            let success := call(gas(), 0, opcode, add(cid, 32), mload(cid), 0x0, 0x400)
            if iszero(success) { revert(0, 0) }
            result := mload(0x0)
        }
    }

    /// @notice Check if a CID exists.
    function exists(bytes memory cid) external returns (bool ok) {
        bytes4 opcode = EXISTS;
        assembly {
            let success := call(gas(), 0, opcode, add(cid, 32), mload(cid), 0x0, 0x20)
            if iszero(success) { revert(0, 0) }
            ok := mload(0x0)
        }
    }

    /// @notice Create a storage entry.
    function storageCreate(bytes memory data) external {
        bytes4 opcode = STORAGE_CREATE;
        assembly {
            let success := call(gas(), 0, opcode, add(data, 32), mload(data), 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }

    /// @notice Query a listing by ID.
    function getListing(bytes32 id) external returns (bytes memory listing) {
        bytes4 opcode = GET_LISTING;
        assembly {
            mstore(0x0, id)
            let success := call(gas(), 0, opcode, 0x0, 0x20, 0x0, 0x400)
            if iszero(success) { revert(0, 0) }
            listing := mload(0x0)
        }
    }

    /// @notice Fetch all listings.
    function listListings() external returns (bytes memory out) {
        bytes4 opcode = LIST_LISTINGS;
        assembly {
            let success := call(gas(), 0, opcode, 0, 0, 0x0, 0x400)
            if iszero(success) { revert(0, 0) }
            out := mload(0x0)
        }
    }

    /// @notice Query a deal by ID.
    function getDeal(bytes32 id) external returns (bytes memory deal) {
        bytes4 opcode = GET_DEAL;
        assembly {
            mstore(0x0, id)
            let success := call(gas(), 0, opcode, 0x0, 0x20, 0x0, 0x400)
            if iszero(success) { revert(0, 0) }
            deal := mload(0x0)
        }
    }

    /// @notice Fetch all deals.
    function listDeals() external returns (bytes memory out) {
        bytes4 opcode = LIST_DEALS;
        assembly {
            let success := call(gas(), 0, opcode, 0, 0, 0x0, 0x400)
            if iszero(success) { revert(0, 0) }
            out := mload(0x0)
        }
    }

}
