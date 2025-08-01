async function loadNavbar() {
  const nav = document.getElementById('nav');
  const res = await fetch('components/navbar.html');
  nav.innerHTML = await res.text();
}

async function loadListings() {
  const res = await fetch('../server/contracts.json').catch(() => ({ json: () => [] }));
  const listings = await res.json();
  const tbody = document.querySelector('#listings tbody');
  if (!tbody) return;
  tbody.innerHTML = listings.map(c => `<tr><td>${c.id}</td><td>${c.name}</td></tr>`).join('');
}

function bindDeploy() {
  const form = document.getElementById('deployForm');
  if (!form) return;
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const name = document.getElementById('name').value;
    const wasm = document.getElementById('wasm').value;
    await fetch('../api/contracts', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, wasm })
    });
    window.location.href = 'listings.html';
  });
}
