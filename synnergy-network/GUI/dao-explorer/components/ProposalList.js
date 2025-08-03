export default function renderProposalList(proposals) {
  const container = document.getElementById("proposal-list");
  container.innerHTML = "";
  proposals.forEach((p) => {
    const item = document.createElement("div");
    item.className = "border p-4 rounded mb-2 bg-white";

    const title = document.createElement("h3");
    title.className = "text-lg font-semibold";
    title.textContent = p.title;

    const description = document.createElement("p");
    description.textContent = p.description;

    const button = document.createElement("button");
    button.className = "text-blue-600 underline";
    button.dataset.id = p.id;
    button.textContent = "View";

    item.append(title, description, button);
    container.appendChild(item);
  });
}
