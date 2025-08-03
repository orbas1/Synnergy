// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title ShardingCoordinator
/// @notice Maintains a mapping of shard identifiers to node addresses.
contract ShardingCoordinator {
    mapping(uint256 => address[]) private shardNodes;

    event NodeAssigned(uint256 indexed shardId, address indexed node);

    /// @notice Assign `node` to manage `shardId`.
    function assignNode(uint256 shardId, address node) external {
        require(node != address(0), "invalid node");
        shardNodes[shardId].push(node);
        emit NodeAssigned(shardId, node);
    }

    /// @notice Retrieve nodes responsible for `shardId`.
    function getNodes(uint256 shardId) external view returns (address[] memory) {
        return shardNodes[shardId];
    }
}
