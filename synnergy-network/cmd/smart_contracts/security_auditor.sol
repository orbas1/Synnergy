// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title SecurityAuditor
/// @notice Records security issues discovered during audits. The contract is
///         intentionally simple and operates independently from the Synnergy
///         VM to provide a clean, testable example.
contract SecurityAuditor {
    address public owner;

    struct Issue {
        string description;
        uint256 timestamp;
    }

    Issue[] private issues;

    event IssueLogged(uint256 indexed index, string description);

    modifier onlyOwner() {
        require(msg.sender == owner, "not authorized");
        _;
    }

    constructor() {
        owner = msg.sender;
    }

    /// @notice Record a new audit finding.
    function logIssue(string calldata description) external onlyOwner {
        issues.push(Issue({description: description, timestamp: block.timestamp}));
        emit IssueLogged(issues.length - 1, description);
    }

    /// @notice Number of recorded issues.
    function totalIssues() external view returns (uint256) {
        return issues.length;
    }

    /// @notice Retrieve an issue by index.
    function getIssue(uint256 index) external view returns (Issue memory) {
        require(index < issues.length, "index out of bounds");
        return issues[index];
    }
}
