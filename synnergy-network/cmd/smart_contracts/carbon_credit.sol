// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title CarbonCredit
/// @notice Minimal ERC20-like token representing certified carbon credits.
contract CarbonCredit {
    string public constant name = "Carbon Credit";
    string public constant symbol = "CO2";
    uint8 public constant decimals = 18;

    uint256 public totalSupply;
    address public admin;
    mapping(address => uint256) public balanceOf;

    event Transfer(address indexed from, address indexed to, uint256 amount);
    event Mint(address indexed to, uint256 amount);

    constructor() {
        admin = msg.sender;
    }

    modifier onlyAdmin() {
        require(msg.sender == admin, "not admin");
        _;
    }

    /// @notice Mint new carbon credits to an account.
    function mint(address to, uint256 amount) external onlyAdmin {
        require(to != address(0), "zero address");
        totalSupply += amount;
        balanceOf[to] += amount;
        emit Mint(to, amount);
        emit Transfer(address(0), to, amount);
    }

    /// @notice Transfer credits to another address.
    function transfer(address to, uint256 amount) external returns (bool) {
        require(balanceOf[msg.sender] >= amount, "insufficient");
        balanceOf[msg.sender] -= amount;
        balanceOf[to] += amount;
        emit Transfer(msg.sender, to, amount);
        return true;
    }
}

