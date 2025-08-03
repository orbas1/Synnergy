// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title ConsensusTester
/// @notice Simple voting contract useful for consensus experiments.
contract ConsensusTester {
    mapping(bytes32 => uint256) public votes;

    event Voted(address indexed voter, bytes32 indexed proposal);

    /// @notice Cast a vote for a proposal identifier.
    function vote(bytes32 proposal) external {
        votes[proposal] += 1;
        emit Voted(msg.sender, proposal);
    }
}

