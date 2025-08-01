// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title DAOExplorer demonstrates Synnergy governance opcodes
/// @notice Uses inline assembly to invoke operations defined in opcode_dispatcher.go
contract DAOExplorer {
    bytes4 constant SUBMIT = 0x0C0005; // SubmitProposal
    bytes4 constant VOTE = 0x0C0007;   // CastVote
    bytes4 constant GET = 0x0C0009;    // GetProposal
    bytes4 constant LIST = 0x0C000A;   // ListProposals

    function submit(bytes memory data) external {
        assembly {
            let success := call(gas(), 0, SUBMIT, add(data, 0x20), mload(data), 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }

    function vote(bytes32 proposalId, bool approve) external {
        let data := mload(0x40)
        mstore(data, proposalId)
        mstore(add(data, 0x20), approve)
        assembly {
            let success := call(gas(), 0, VOTE, data, 0x40, 0, 0)
            if iszero(success) { revert(0, 0) }
        }
    }

    function get(bytes32 proposalId) external returns (bytes memory result) {
        let data := mload(0x40)
        mstore(data, proposalId)
        assembly {
            let success := call(gas(), 0, GET, data, 0x20, 0x100, 0x400)
            if iszero(success) { revert(0, 0) }
            result := mload(0x100)
        }
    }

    function list() external returns (bytes memory result) {
        assembly {
            let success := call(gas(), 0, LIST, 0, 0, 0x100, 0x400)
            if iszero(success) { revert(0, 0) }
            result := mload(0x100)
        }
    }
}
