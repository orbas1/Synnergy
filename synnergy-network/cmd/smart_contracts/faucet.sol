pragma solidity ^0.8.20;

/// @title Faucet Contract
/// @notice Simple token faucet with rate limiting using Synnergy opcodes.
contract Faucet {
    uint64 public dripAmount;
    uint256 public cooldown;
    mapping(address => uint256) public nextRequestTime;

    constructor(uint64 _dripAmount, uint256 _cooldown) {
        dripAmount = _dripAmount;
        cooldown = _cooldown;
    }

    /// Deposit tokens into the faucet for distribution.
    function deposit(uint32 tokenId, uint64 amount) external {
        bytes4 op = 0x190004; // Tokens_Transfer
        address fa = address(this);
        assembly {
            mstore(0x0, tokenId)
            mstore(0x20, caller())
            mstore(0x40, fa)
            mstore(0x60, amount)
            if iszero(call(gas(), 0, op, 0x0, 0x80, 0, 0)) { revert(0,0) }
        }
    }

    /// Request tokens from the faucet respecting the cooldown.
    function request(uint32 tokenId) external {
        require(block.timestamp >= nextRequestTime[msg.sender], "cooldown");
        nextRequestTime[msg.sender] = block.timestamp + cooldown;

        bytes4 op = 0x190004; // Tokens_Transfer
        address fa = address(this);
        uint64 amount = dripAmount;
        assembly {
            mstore(0x0, tokenId)
            mstore(0x20, fa)
            mstore(0x40, caller())
            mstore(0x60, amount)
            if iszero(call(gas(), 0, op, 0x0, 0x80, 0, 0)) { revert(0,0) }
        }
    }
}
