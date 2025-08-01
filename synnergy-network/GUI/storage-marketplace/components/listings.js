const API = '/api/listings';

export async function renderListings() {
  const res = await fetch(API);
  const data = await res.json();
  const tbody = document.querySelector('#listingsTable tbody');
  tbody.innerHTML = '';
  data.forEach(l => {
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td>${l.id}</td>
      <td>${l.provider}</td>
      <td>${l.pricePerGB}</td>
      <td>${l.capacityGB}</td>
      <td>${new Date(l.createdAt).toLocaleString()}</td>
    `;
    tbody.appendChild(tr);
  });
}

export function initListingForm() {
  const form = document.getElementById('listingForm');
  form.addEventListener('submit', async e => {
    e.preventDefault();
    const data = Object.fromEntries(new FormData(form).entries());
    await fetch(API, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    });
    form.reset();
    renderListings();
  });
}
