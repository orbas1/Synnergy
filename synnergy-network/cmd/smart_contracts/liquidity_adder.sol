// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title LiquidityAdder showcases interaction with Synnergy's liquidity pool
///        using opcode 0x0F0004 (Liquidity_AddLiquidity).
/// @notice The gas table assigns 5,000 gas to this opcode.
contract LiquidityAdder {
    /// Add liquidity to a pool via a custom VM opcode.
    /// @param poolId Identifier of the pool
    /// @param amount Amount of base asset provided
    function add(uint32 poolId, uint64 amount) external {
        bytes4 opcode = 0x0F0004; // Liquidity_AddLiquidity
        assembly {
            mstore(0x0, poolId)
            mstore(0x20, amount)
            let success := call(gas(), 0, opcode, 0x0, 0x40, 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }
}
