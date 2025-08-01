// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title TokenFactory demonstrates multiple Synnergy opcodes
/// @notice Uses inline assembly to directly invoke VM instructions
///         from opcode_dispatcher.go for token management
contract TokenFactory {
    /// Deploy a new token and optionally mint initial supply to caller
    function createToken(uint8 standard, uint8 decimals, bool fixed, uint64 supply) external returns (uint32 tokenId) {
        bytes4 createOp = 0x190010; // Tokens_Create
        bytes4 mintOp = 0x1C001A;   // MintToken_VM
        assembly {
            mstore(0x0, standard)
            mstore(0x20, decimals)
            mstore(0x40, fixed)
            mstore(0x60, supply)
            let success := call(gas(), 0, createOp, 0x0, 0x80, 0x0, 0x20)
            if iszero(success) { revert(0, 0) }
            tokenId := mload(0x0)
            if gt(supply, 0) {
                mstore(0x20, caller())
                mstore(0x40, supply)
                success := call(gas(), 0, mintOp, 0x0, 0x60, 0, 0)
                if iszero(success) { revert(0, 0) }
            }
        }
    }

    /// Mint additional tokens to an address
    function mint(uint32 tokenId, address to, uint64 amount) external {
        bytes4 op = 0x1C001A; // MintToken_VM
        assembly {
            mstore(0x0, tokenId)
            mstore(0x20, to)
            mstore(0x40, amount)
            if iszero(call(gas(), 0, op, 0x0, 0x60, 0, 0)) { revert(0, 0) }
        }
    }

    /// Transfer tokens using the Tokens_Transfer opcode
    function transfer(uint32 tokenId, address to, uint64 amount) external {
        bytes4 op = 0x190004; // Tokens_Transfer
        assembly {
            mstore(0x0, tokenId)
            mstore(0x20, caller())
            mstore(0x40, to)
            mstore(0x60, amount)
            if iszero(call(gas(), 0, op, 0x0, 0x80, 0, 0)) { revert(0, 0) }
        }
    }

    /// Burn caller's tokens
    function burn(uint32 tokenId, uint64 amount) external {
        bytes4 op = 0x190008; // Tokens_Burn
        assembly {
            mstore(0x0, tokenId)
            mstore(0x20, caller())
            mstore(0x40, amount)
            if iszero(call(gas(), 0, op, 0x0, 0x60, 0, 0)) { revert(0, 0) }
        }
    }

    /// Query balance using Tokens_BalanceOf
    function balanceOf(uint32 tokenId, address owner) external returns (uint64 bal) {
        bytes4 op = 0x190003; // Tokens_BalanceOf
        assembly {
            mstore(0x0, tokenId)
            mstore(0x20, owner)
            if iszero(call(gas(), 0, op, 0x0, 0x40, 0x0, 0x20)) { revert(0, 0) }
            bal := mload(0x0)
        }
    }
}
