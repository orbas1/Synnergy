// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title AI Optimizer
/// @author Synnergy Network
/// @notice Tracks the best optimization score submitted by participants.
contract AIOptimizer {
    struct Result {
        uint256 score;          // Optimization score (higher is better)
        address submitter;      // Address that achieved the score
    }

    /// @notice Current best result.
    Result public best;

    /// @notice Emitted when a new best score is submitted.
    /// @param score The score that beat the previous best.
    /// @param submitter Address that provided the optimization.
    event NewBest(uint256 indexed score, address indexed submitter);

    /// @notice Submit a new optimization score.
    /// @param score The achieved optimization score.
    function submit(uint256 score) external {
        if (score > best.score) {
            best = Result(score, msg.sender);
            emit NewBest(score, msg.sender);
        }
    }
}
