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

    /// Purchase and deploy the contract
    function purchase(uint256 id) external {
        Listing storage l = listings[id];
        require(l.price > 0, "invalid listing");
        assembly {
            let opcode := 0x080004 // Deploy
            // layout [wasm_ptr, wasm_len]
            let ptr := add(l, 0x40)
            mstore(0x0, ptr)
            mstore(0x20, mload(ptr))
            let success := call(gas(), 0, opcode, 0x0, 0x40, 0, 0)
            if iszero(success) { revert(0,0) }
        }
        delete listings[id];
    }
}
