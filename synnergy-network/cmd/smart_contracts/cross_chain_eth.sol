// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

// CrossChainBridge demonstrates a minimal bridge contract that
// locks Ether on this chain and emits events for off-chain relayers.
// Admins can release funds after verifying proofs from the remote chain.
// Some verification logic uses inline assembly to showcase opcode-level
// operations.

contract CrossChainBridge {
    address public admin;
    mapping(address => uint256) public balances;

    event Deposit(address indexed from, bytes32 indexed toChain, uint256 amount);
    event Withdraw(address indexed to, uint256 amount);

    modifier onlyAdmin() {
        require(msg.sender == admin, "not admin");
        _;
    }

    constructor() {
        admin = msg.sender;
    }

    // Lock Ether and emit a deposit event containing the destination
    // chain address encoded as bytes32.
    function deposit(bytes32 toChainAddr) external payable {
        require(msg.value > 0, "zero value");
        balances[msg.sender] += msg.value;
        emit Deposit(msg.sender, toChainAddr, msg.value);
    }

    // Release funds to a recipient after verifying a proof from the
    // remote chain. The proof format is application specific and is
    // validated with inline assembly for illustration.
    function withdraw(address payable to, uint256 amount, bytes calldata proof) external onlyAdmin {
        require(balances[to] >= amount, "insufficient");
        require(_verifyProof(proof), "bad proof");
        balances[to] -= amount;
        emit Withdraw(to, amount);
        (bool ok, ) = to.call{value: amount}("");
        require(ok, "transfer failed");
    }

    // Example proof verification using assembly. For this demo we simply
    // check that the first byte of the proof is 0x01, but more complex
    // logic could be implemented directly with opcodes.
    function _verifyProof(bytes calldata proof) private pure returns (bool valid) {
        assembly {
            // proof.offset gives the memory position of the bytes data
            valid := eq(byte(0, calldataload(proof.offset)), 0x01)
        }
    }
}
