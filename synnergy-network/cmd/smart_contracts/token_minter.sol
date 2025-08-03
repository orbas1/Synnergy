// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title TokenMinter demonstrates calling Synnergy's custom MintToken opcode
/// @notice Uses inline assembly to invoke opcode 0x1C001A defined in
///         `opcode_dispatcher.go`. Gas cost is 2,000 per `gas_table.go`.
contract TokenMinter {
    /// Mint tokens to a recipient using a Synnergy VM opcode.
    /// @param tokenId The 32-bit asset identifier
    /// @param recipient Destination address that receives the minted amount
    /// @param amount Number of tokens to mint
    function mint(uint32 tokenId, address recipient, uint64 amount) external {
        bytes4 opcode = bytes4(0x001C001A); // MintToken_VM
        assembly {
            // Layout: [tokenId(32)][recipient(32)][amount(32)]
            mstore(0x0, tokenId)
            mstore(0x20, recipient)
            mstore(0x40, amount)
            // Call precompile-like address 0 with custom opcode
            let success := call(gas(), 0, opcode, 0x0, 0x60, 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }
}
