// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title CrossChainProtocol
/// @notice Emits events representing cross-chain messages. The actual message
///         passing logic is expected to be handled by off-chain relayers.
contract CrossChainProtocol {
    event MessageSent(bytes32 indexed dstChain, address indexed from, address to, bytes data);
    event MessageReceived(bytes32 indexed srcChain, address indexed from, address to, bytes data);

    /// @notice Send a message to a remote chain.
    function sendMessage(bytes32 dstChain, address to, bytes calldata data) external {
        emit MessageSent(dstChain, msg.sender, to, data);
    }

    /// @notice Receive a message from a remote chain.
    function receiveMessage(bytes32 srcChain, address from, address to, bytes calldata data) external {
        emit MessageReceived(srcChain, from, to, data);
    }
}

