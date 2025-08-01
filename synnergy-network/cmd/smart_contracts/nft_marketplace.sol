// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title NFTMarketplace example contract for Synnergy Network
/// @notice Demonstrates listing and purchasing NFTs via Synnergy opcodes.
contract NFTMarketplace {
    struct Listing {
        uint32 tokenId;
        address owner;
        uint64 price;
        bool active;
    }

    mapping(uint256 => Listing) public listings;
    uint256 public nextId;

    bytes4 constant OP_BALANCE = 0x190003;  // Tokens_BalanceOf
    bytes4 constant OP_TRANSFER = 0x190004; // Tokens_Transfer

    /// List an NFT for sale
    function list(uint32 tokenId, uint64 price) external {
        assembly {
            // Verify caller owns at least one of tokenId
            mstore(0x0, tokenId)
            mstore(0x20, caller())
            let ok := call(gas(), 0, OP_BALANCE, 0x0, 0x40, 0x0, 0x20)
            if iszero(ok) { revert(0,0) }
            // zero means no balance
            if iszero(mload(0x0)) { revert(0,0) }
        }
        listings[nextId] = Listing(tokenId, msg.sender, price, true);
        nextId++;
    }

    /// Purchase a listed NFT
    function buy(uint256 id) external {
        Listing storage l = listings[id];
        require(l.active, "inactive");
        assembly {
            mstore(0x0, l.tokenId)
            mstore(0x20, caller())
            mstore(0x40, l.owner)
            mstore(0x60, l.price)
            let ok := call(gas(), 0, OP_TRANSFER, 0x0, 0x80, 0, 0)
            if iszero(ok) { revert(0,0) }
        }
        l.active = false;
    }
}
