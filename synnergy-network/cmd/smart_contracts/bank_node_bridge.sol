// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title BankNodeBridge
/// @notice Example bridge contract handling ERC20 token transfers between a
///         bank custody system and the blockchain.
interface IERC20 {
    function transfer(address to, uint256 amount) external returns (bool);
    function transferFrom(address from, address to, uint256 amount) external returns (bool);
}

contract BankNodeBridge {
    IERC20 public immutable token;
    address public admin;
    mapping(address => uint256) public balances;

    event Deposited(address indexed from, uint256 amount);
    event Withdrawn(address indexed to, uint256 amount);

    constructor(IERC20 _token) {
        token = _token;
        admin = msg.sender;
    }

    modifier onlyAdmin() {
        require(msg.sender == admin, "not admin");
        _;
    }

    /// @notice Lock tokens from the user and credit their bridge balance.
    function deposit(uint256 amount) external {
        require(amount > 0, "zero amount");
        token.transferFrom(msg.sender, address(this), amount);
        balances[msg.sender] += amount;
        emit Deposited(msg.sender, amount);
    }

    /// @notice Release tokens to a recipient after off-chain approval.
    function withdraw(address to, uint256 amount) external onlyAdmin {
        require(balances[to] >= amount, "insufficient");
        balances[to] -= amount;
        token.transfer(to, amount);
        emit Withdrawn(to, amount);
    }
}

