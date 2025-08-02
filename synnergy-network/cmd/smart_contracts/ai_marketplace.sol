// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title AI Service Marketplace Contract
/// @author Synnergy Network
/// @notice Demonstrates how Synnergy VM opcodes can be used for
///         a decentralized AI service marketplace.
contract AIServiceMarketplace {
    struct Service {
        string name;
        uint256 price;
        address provider;
    }

    /// @notice Registered services keyed by identifier.
    mapping(uint256 => Service) public services;
    /// @notice Next service identifier.
    uint256 public nextId;

    /// @notice Emitted when a service is registered.
    /// @param id Identifier for the service.
    /// @param name Name of the service.
    /// @param price Cost in wei for the service.
    /// @param provider Address of the service provider.
    event ServiceRegistered(
        uint256 indexed id,
        string name,
        uint256 indexed price,
        address indexed provider
    );

    /// @notice Emitted when a service is purchased.
    /// @param id Identifier for the purchased service.
    /// @param buyer Address that purchased the service.
    event ServicePurchased(uint256 indexed id, address indexed buyer);

    /// @dev Error thrown when a service identifier does not exist.
    error ServiceNotFound();

    /// @dev Error thrown when the payment sent does not match the service price.
    error IncorrectPrice();

    /// Register a new AI service available for purchase.
    /// @notice Register a new AI service available for purchase.
    /// @param name Name of the service.
    /// @param price Cost in wei for the service.
    function registerService(string calldata name, uint256 price) external {
        uint256 id = nextId;
        ++nextId;
        services[id] = Service(name, price, msg.sender);
        emit ServiceRegistered(id, name, price, msg.sender);
    }

    /// @notice Purchase a service. Payment is transferred using a Synnergy opcode.
    /// @param id Identifier of the service to purchase.
    function buyService(uint256 id) external payable {
        Service memory svc = services[id];
        if (bytes(svc.name).length == 0) revert ServiceNotFound();
        if (msg.value != svc.price) revert IncorrectPrice();
        _transfer(svc.provider, msg.value);
        emit ServicePurchased(id, msg.sender);
    }

    /// @notice Low-level ETH transfer using VM opcode `0x1C001B` (VM_Transfer).
    /// @param to Recipient address.
    /// @param amount Amount of wei to transfer.
    function _transfer(address to, uint256 amount) private {
        // Explicitly cast the opcode to `bytes4` to satisfy the type checker.
        bytes4 opcode = bytes4(uint32(0x001C001B)); // VM_Transfer
        // solhint-disable-next-line no-inline-assembly
        assembly {
            // Layout: [to(32)][amount(32)]
            mstore(0x0, to)
            mstore(0x20, amount)
            let success := call(gas(), 0, opcode, 0x0, 0x40, 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }
}
