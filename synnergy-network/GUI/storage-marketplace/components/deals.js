const API = '/api/deals';

export async function renderDeals() {
  const res = await fetch(API);
  const data = await res.json();
  const tbody = document.querySelector('#dealsTable tbody');
  tbody.innerHTML = '';
  data.forEach(d => {
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td>${d.id}</td>
      <td>${d.listingId}</td>
      <td>${d.client}</td>
      <td>${d.duration}</td>
      <td>${new Date(d.createdAt).toLocaleString()}</td>
    `;
    tbody.appendChild(tr);
  });
}

export function initDealForm() {
  const form = document.getElementById('dealForm');
  form.addEventListener('submit', async e => {
    e.preventDefault();
    const data = Object.fromEntries(new FormData(form).entries());
    await fetch(API, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    });
    form.reset();
    renderDeals();
  });
}
