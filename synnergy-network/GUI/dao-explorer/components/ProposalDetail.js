export default function renderProposalDetail(proposal) {
  const container = document.getElementById("proposal-detail");
  container.innerHTML = "";

  if (!proposal) {
    const placeholder = document.createElement("p");
    placeholder.textContent = "Select a proposal";
    container.appendChild(placeholder);
    return;
  }

  const title = document.createElement("h2");
  title.className = "text-xl font-bold mb-2";
  title.textContent = proposal.title;

  const description = document.createElement("p");
  description.className = "mb-2";
  description.textContent = proposal.description;

  const votes = document.createElement("p");
  votes.textContent = `For: ${proposal.votesFor} | Against: ${proposal.votesAgainst}`;

  const form = document.createElement("form");
  form.id = "vote-form";
  form.className = "mt-2";

  const hiddenId = document.createElement("input");
  hiddenId.type = "hidden";
  hiddenId.name = "id";
  hiddenId.value = proposal.id;

  const approveBtn = document.createElement("button");
  approveBtn.name = "approve";
  approveBtn.value = "true";
  approveBtn.className = "bg-green-500 text-white px-2 py-1 mr-2 rounded";
  approveBtn.textContent = "Approve";

  const rejectBtn = document.createElement("button");
  rejectBtn.name = "approve";
  rejectBtn.value = "false";
  rejectBtn.className = "bg-red-500 text-white px-2 py-1 rounded";
  rejectBtn.textContent = "Reject";

  form.append(hiddenId, approveBtn, rejectBtn);
  container.append(title, description, votes, form);
}
