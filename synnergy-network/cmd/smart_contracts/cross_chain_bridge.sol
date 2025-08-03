// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title CrossChainTokenBridge
/// @notice Locks ERC20 tokens and emits events for off-chain relayers who
///         mint or release assets on a remote chain.
interface IERC20 {
    function transfer(address to, uint256 amount) external returns (bool);
    function transferFrom(address from, address to, uint256 amount) external returns (bool);
}

contract CrossChainTokenBridge {
    IERC20 public immutable token;
    address public admin;
    mapping(address => uint256) public balances;

    event Deposit(address indexed from, bytes32 indexed toChain, uint256 amount);
    event Withdraw(address indexed to, uint256 amount);

    constructor(IERC20 _token) {
        token = _token;
        admin = msg.sender;
    }

    modifier onlyAdmin() {
        require(msg.sender == admin, "not admin");
        _;
    }

    /// @notice Lock tokens and emit a deposit event.
    function deposit(uint256 amount, bytes32 toChainAddr) external {
        require(amount > 0, "zero amount");
        token.transferFrom(msg.sender, address(this), amount);
        balances[msg.sender] += amount;
        emit Deposit(msg.sender, toChainAddr, amount);
    }

    /// @notice Release tokens after verifying an off-chain proof.
    function withdraw(address to, uint256 amount, bytes calldata proof) external onlyAdmin {
        require(balances[to] >= amount, "insufficient");
        require(_verifyProof(proof), "bad proof");
        balances[to] -= amount;
        token.transfer(to, amount);
        emit Withdraw(to, amount);
    }

    // Example proof verification using assembly; checks first byte == 0x01.
    function _verifyProof(bytes calldata proof) private pure returns (bool valid) {
        assembly {
            valid := eq(byte(0, calldataload(proof.offset)), 0x01)
        }
    }
}

