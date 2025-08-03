// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title DAOExplorer demonstrates Synnergy governance opcodes
/// @notice Uses inline assembly to invoke operations defined in opcode_dispatcher.go
contract DAOExplorer {
    // Precompile addresses for governance operations
    uint256 constant SUBMIT = 0x0C0005;        // SubmitProposal
    uint256 constant VOTE = 0x0C0007;          // CastVote
    uint256 constant GET = 0x0C0009;           // GetProposal
    uint256 constant LIST = 0x0C000A;          // ListProposals
    uint256 constant EXECUTE = 0x0C0008;       // ExecuteProposal
    uint256 constant BALANCE_OF = 0x0C0006;    // BalanceOfAsset

    function submit(bytes memory data) external {
        assembly {
            let success := call(gas(), SUBMIT, 0, add(data, 0x20), mload(data), 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }

    function vote(bytes32 proposalId, bool approve) external {
        assembly {
            let data := mload(0x40)
            mstore(data, proposalId)
            mstore(add(data, 0x20), approve)
            let success := call(gas(), VOTE, 0, data, 0x40, 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }

    function get(bytes32 proposalId) external returns (bytes memory result) {
        assembly {
            let dataPtr := mload(0x40)
            mstore(dataPtr, proposalId)
            let success := call(gas(), GET, 0, dataPtr, 0x20, 0, 0)
            if iszero(success) { revert(0, 0) }

            let size := returndatasize()
            result := mload(0x40)
            mstore(result, size)
            let copyPtr := add(result, 0x20)
            returndatacopy(copyPtr, 0, size)
            mstore(0x40, add(copyPtr, size))
        }
    }

    function list() external returns (bytes memory result) {
        assembly {
            let success := call(gas(), LIST, 0, 0, 0, 0, 0)
            if iszero(success) { revert(0, 0) }

            let size := returndatasize()
            result := mload(0x40)
            mstore(result, size)
            let ptr := add(result, 0x20)
            returndatacopy(ptr, 0, size)
            mstore(0x40, add(ptr, size))
        }
    }

    function execute(bytes32 proposalId) external {
        assembly {
            let dataPtr := mload(0x40)
            mstore(dataPtr, proposalId)
            let success := call(gas(), EXECUTE, 0, dataPtr, 0x20, 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }

    function balanceOf(address asset) external returns (uint256 bal) {
        assembly {
            let dataPtr := mload(0x40)
            mstore(dataPtr, asset)
            let outPtr := add(dataPtr, 0x20)
            let success := call(gas(), BALANCE_OF, 0, dataPtr, 0x20, outPtr, 0x20)
            if iszero(success) { revert(0, 0) }
            bal := mload(outPtr)
            mstore(0x40, add(outPtr, 0x20))
        }
    }
}
