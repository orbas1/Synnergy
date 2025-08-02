/**
 * Proposal service. If a contract address is configured this module will
 * interact with the DAOExplorer Solidity contract using ethers.js. Otherwise it
 * falls back to an in-memory store for demonstration purposes.
 */
const config = require("../config/config");
let contract;
let proposals = new Map();
let idCounter = 1;

if (config.contractAddress) {
  // Lazy load contractService only when needed to keep environment setup simple
  contract = require("./contractService");
}

async function listProposals() {
  if (contract) {
    return await contract.listProposals();
  }
  return Array.from(proposals.values());
}

async function getProposal(id) {
  if (contract) {
    return await contract.getProposal(id);
  }
  return proposals.get(id);
}

async function submitProposal(data) {
  if (contract) {
    const txHash = await contract.submitProposal(data);
    return { id: txHash, ...data };
  }
  const id = String(idCounter++);
  const proposal = { id, ...data, votesFor: 0, votesAgainst: 0 };
  proposals.set(id, proposal);
  return proposal;
}

async function castVote(id, approve) {
  if (contract) {
    await contract.vote(id, approve);
    return getProposal(id);
  }
  const proposal = proposals.get(id);
  if (!proposal) return null;
  if (approve) {
    proposal.votesFor += 1;
  } else {
    proposal.votesAgainst += 1;
  }
  return proposal;
}

async function execute(id) {
  if (contract) {
    return contract.executeProposal(id);
  }
  return null;
}

async function balance(id, asset) {
  if (contract) {
    return contract.balanceOfAsset(asset);
  }
  return null;
}

module.exports = {
  listProposals,
  getProposal,
  submitProposal,
  castVote,
  execute,
  balance,
};
