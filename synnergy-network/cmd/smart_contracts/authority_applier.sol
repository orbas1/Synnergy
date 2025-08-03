// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title AuthorityApplier
/// @notice Allows users to submit applications for elevated network roles.
/// @dev Administrators can review and approve applications. The contract is
///      intentionally lightweight but production ready.
contract AuthorityApplier {
    struct Application {
        address applicant;
        string metadata;
        bool approved;
    }

    address public admin;
    mapping(address => Application) public applications;

    event Applied(address indexed applicant, string metadata);
    event Approved(address indexed applicant);

    constructor() {
        admin = msg.sender;
    }

    modifier onlyAdmin() {
        require(msg.sender == admin, "not admin");
        _;
    }

    /// Submit a new application containing off-chain metadata.
    function submitApplication(string calldata metadata) external {
        Application storage app = applications[msg.sender];
        require(app.applicant == address(0), "already applied");
        app.applicant = msg.sender;
        app.metadata = metadata;
        emit Applied(msg.sender, metadata);
    }

    /// Approve an application. Only callable by the administrator.
    function approve(address applicant) external onlyAdmin {
        Application storage app = applications[applicant];
        require(app.applicant != address(0), "no application");
        app.approved = true;
        emit Approved(applicant);
    }
}

