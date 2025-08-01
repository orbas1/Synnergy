/**
 * Simulated proposal storage. In real implementation this would
 * call Synnergy smart contracts using opcodes defined in opcode_dispatcher.go
 * such as SubmitProposal, CastVote and ListProposals.
 */
const proposals = new Map();
let idCounter = 1;

function listProposals() {
  return Array.from(proposals.values());
}

function getProposal(id) {
  return proposals.get(id);
}

function submitProposal(data) {
  const id = String(idCounter++);
  const proposal = { id, ...data, votesFor: 0, votesAgainst: 0 };
  proposals.set(id, proposal);
  return proposal;
}

function castVote(id, approve) {
  const proposal = proposals.get(id);
  if (!proposal) return null;
  if (approve) {
    proposal.votesFor += 1;
  } else {
    proposal.votesAgainst += 1;
  }
  return proposal;
}

module.exports = {
  listProposals,
  getProposal,
  submitProposal,
  castVote
};
