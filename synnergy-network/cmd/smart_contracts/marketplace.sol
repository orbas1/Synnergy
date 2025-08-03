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

    /// @notice Deploy a WASM contract using Synnergy opcode 0x080004 (Deploy)
    /// @param wasm WASM bytecode of the contract
    /// @param price Asking price in SYNN tokens
    /// @return id Identifier of the newly created listing
    function list(bytes calldata wasm, uint256 price) external returns (uint256 id) {
        id = nextId++;
        listings[id] = Listing(msg.sender, price, wasm);
    }

    /// @notice Purchase and deploy the contract using Synnergy opcodes
    /// @param id Listing identifier
    function purchase(uint256 id) external {
        Listing storage l = listings[id];
        require(l.price > 0, "invalid listing");

        address seller = l.seller;
        uint256 price = l.price;
        bytes memory code = l.wasm;

        assembly {
            // transfer SYNN tokens from buyer to seller
            let transfer := 0x190004 // Tokens_Transfer
            mstore(0x0, caller())
            mstore(0x20, seller)
            mstore(0x40, price)
            if iszero(call(gas(), transfer, 0, 0x0, 0x60, 0, 0)) { revert(0, 0) }

            // deploy WASM contract
            let deploy := 0x080004 // Deploy
            let ptr := add(code, 0x20)
            mstore(0x0, ptr)
            mstore(0x20, mload(code))
            if iszero(call(gas(), deploy, 0, 0x0, 0x40, 0, 0)) { revert(0, 0) }
        }

        delete listings[id];
    }
}
