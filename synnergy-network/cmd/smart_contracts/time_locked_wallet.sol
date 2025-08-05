// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title TimeLockedWallet
/// @notice Ether deposited into this contract can only be withdrawn after the
///         configured `releaseTime`.
contract TimeLockedWallet {
    address public owner;
    uint256 public releaseTime;

    constructor(uint256 _releaseTime) {
        require(_releaseTime > block.timestamp, "release time in past");
        owner = msg.sender;
        releaseTime = _releaseTime;
    }

    /// @notice Accept ether deposits.
    receive() external payable {}

    /// @notice Withdraw all ether after `releaseTime` has passed.
    function withdraw() external {
        require(msg.sender == owner, "not owner");
        require(block.timestamp >= releaseTime, "locked");
        payable(owner).transfer(address(this).balance);
    }
}
