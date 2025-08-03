// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title AMMPoolManager
/// @notice Minimal constant product automated market maker for two ERC20 tokens.
/// @dev This contract is a simplified example and does not issue LP tokens.
interface IERC20 {
    function transfer(address to, uint256 amount) external returns (bool);
    function transferFrom(address from, address to, uint256 amount) external returns (bool);
    function balanceOf(address owner) external view returns (uint256);
}

contract AMMPoolManager {
    IERC20 public immutable token0;
    IERC20 public immutable token1;
    uint256 public reserve0;
    uint256 public reserve1;

    event LiquidityAdded(address indexed provider, uint256 amount0, uint256 amount1);
    event Swapped(address indexed trader, uint256 amount0Out, uint256 amount1Out);

    constructor(IERC20 _token0, IERC20 _token1) {
        token0 = _token0;
        token1 = _token1;
    }

    /// @notice Add balanced liquidity to the pool.
    function addLiquidity(uint256 amount0, uint256 amount1) external {
        require(amount0 > 0 && amount1 > 0, "zero amount");
        token0.transferFrom(msg.sender, address(this), amount0);
        token1.transferFrom(msg.sender, address(this), amount1);
        reserve0 += amount0;
        reserve1 += amount1;
        emit LiquidityAdded(msg.sender, amount0, amount1);
    }

    /// @notice Swap one token for the other using the constant product formula.
    function swap(uint256 amount0Out, uint256 amount1Out, address to) external {
        require(amount0Out > 0 || amount1Out > 0, "no output");
        require(amount0Out < reserve0 && amount1Out < reserve1, "liquidity");
        if (amount0Out > 0) token0.transfer(to, amount0Out);
        if (amount1Out > 0) token1.transfer(to, amount1Out);

        uint256 balance0 = token0.balanceOf(address(this));
        uint256 balance1 = token1.balanceOf(address(this));
        uint256 amount0In = balance0 > reserve0 - amount0Out ? balance0 - (reserve0 - amount0Out) : 0;
        uint256 amount1In = balance1 > reserve1 - amount1Out ? balance1 - (reserve1 - amount1Out) : 0;
        require(amount0In > 0 || amount1In > 0, "insufficient input");
        // invariant check: (reserve0 + amount0In - amount0Out) * (reserve1 + amount1In - amount1Out) >= reserve0 * reserve1
        require((balance0 * balance1) >= (reserve0 * reserve1), "K");
        reserve0 = balance0;
        reserve1 = balance1;
        emit Swapped(msg.sender, amount0Out, amount1Out);
    }
}

