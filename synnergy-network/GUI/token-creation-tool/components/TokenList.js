export async function renderTokenList(container) {
  const res = await fetch('/api/tokens');
  const tokens = await res.json();
  container.innerHTML = `
    <h2 class="h4 mt-5">Existing Tokens</h2>
    <table class="table table-striped">
      <thead><tr><th>ID</th><th>Name</th><th>Symbol</th><th>Supply</th></tr></thead>
      <tbody>${tokens.map(t => `
        <tr><td>${t.id}</td><td>${t.name}</td><td>${t.symbol}</td><td>${t.supply}</td></tr>
      `).join('')}</tbody>
    </table>`;
}
