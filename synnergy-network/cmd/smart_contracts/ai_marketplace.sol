// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title AI Service Marketplace Contract
/// @notice Demonstrates how Synnergy VM opcodes can be used for
///         a decentralized AI service marketplace.
contract AIServiceMarketplace {
    struct Service {
        string name;
        uint256 price;
        address provider;
    }

    mapping(uint256 => Service) public services;
    uint256 public nextId;

    event ServiceRegistered(uint256 indexed id, string name, uint256 price, address provider);
    event ServicePurchased(uint256 indexed id, address buyer);

    /// Register a new AI service available for purchase.
    function registerService(string calldata name, uint256 price) external {
        uint256 id = nextId++;
        services[id] = Service(name, price, msg.sender);
        emit ServiceRegistered(id, name, price, msg.sender);
    }

    /// Purchase a service. Payment is transferred using a Synnergy opcode.
    function buyService(uint256 id) external payable {
        Service memory svc = services[id];
        require(bytes(svc.name).length != 0, "not found");
        require(msg.value == svc.price, "bad price");
        _transfer(svc.provider, msg.value);
        emit ServicePurchased(id, msg.sender);
    }

    /// Low-level ETH transfer using VM opcode 0x1C001B (VM_Transfer).
    function _transfer(address to, uint256 amount) private {
        bytes4 opcode = 0x1C001B; // VM_Transfer
        assembly {
            // Layout: [to(32)][amount(32)]
            mstore(0x0, to)
            mstore(0x20, amount)
            let success := call(gas(), 0, opcode, 0x0, 0x40, 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }
}
