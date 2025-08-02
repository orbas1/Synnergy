export default function renderProposalDetail(proposal) {
  const container = document.getElementById("proposal-detail");
  container.innerHTML = proposal
    ? `
    <h2 class="text-xl font-bold mb-2">${proposal.title}</h2>
    <p class="mb-2">${proposal.description}</p>
    <p>For: ${proposal.votesFor} | Against: ${proposal.votesAgainst}</p>
    <form id="vote-form" class="mt-2">
      <input type="hidden" name="id" value="${proposal.id}">
      <button name="approve" value="true" class="bg-green-500 text-white px-2 py-1 mr-2 rounded">Approve</button>
      <button name="approve" value="false" class="bg-red-500 text-white px-2 py-1 rounded">Reject</button>
    </form>
  `
    : "<p>Select a proposal</p>";
}
