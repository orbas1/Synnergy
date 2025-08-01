// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title Simple Marketplace Contract
/// @notice Demonstrates deployment and execution using Synnergy opcodes.
contract Marketplace {
    struct Listing {
        address seller;
        uint256 price;
        bytes wasm;
    }

    mapping(uint256 => Listing) public listings;
    uint256 public nextId;

    /// Deploy a WASM contract using Synnergy opcode 0x080004 (Deploy)
    function list(bytes calldata wasm, uint256 price) external returns (uint256) {
        uint256 id = nextId++;
        listings[id] = Listing(msg.sender, price, wasm);
        return id;
    }

    /// Purchase and deploy the contract using Synnergy opcodes
    function purchase(uint256 id) external {
        Listing storage l = listings[id];
        require(l.price > 0, "invalid listing");
        assembly {
            // transfer SYNN tokens from buyer to seller
            let transfer := 0x190004 // Tokens_Transfer
            mstore(0x0, caller())
            mstore(0x20, l.seller)
            mstore(0x40, l.price)
            if iszero(call(gas(), 0, transfer, 0x0, 0x60, 0, 0)) { revert(0,0) }

            // deploy WASM contract
            let deploy := 0x080004 // Deploy
            let ptr := add(l, 0x40)
            mstore(0x0, ptr)
            mstore(0x20, mload(ptr))
            if iszero(call(gas(), 0, deploy, 0x0, 0x40, 0, 0)) { revert(0,0) }
        }
        delete listings[id];
    }
}
