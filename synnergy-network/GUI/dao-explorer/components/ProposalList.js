export default function renderProposalList(proposals) {
  const container = document.getElementById('proposal-list');
  container.innerHTML = '';
  proposals.forEach(p => {
    const item = document.createElement('div');
    item.className = 'border p-4 rounded mb-2 bg-white';
    item.innerHTML = `<h3 class="text-lg font-semibold">${p.title}</h3>
      <p>${p.description}</p>
      <button class="text-blue-600 underline" data-id="${p.id}">View</button>`;
    container.appendChild(item);
  });
}
