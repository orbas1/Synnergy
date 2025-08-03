export async function renderTokenList(container) {
  try {
    const res = await fetch("/api/tokens");
    if (!res.ok) {
      throw new Error(`Request failed with status ${res.status}`);
    }
    const tokens = await res.json();
    container.innerHTML = `
    <h2 class="h4 mt-5">Existing Tokens</h2>
    <table class="table table-striped">
      <thead><tr><th>ID</th><th>Name</th><th>Symbol</th><th>Supply</th></tr></thead>
      <tbody>${tokens
        .map(
          (t) => `
        <tr><td>${t.id}</td><td>${t.name}</td><td>${t.symbol}</td><td>${t.supply}</td></tr>
      `
        )
        .join("")}</tbody>
    </table>`;
  } catch (err) {
    container.innerHTML = `<p class="text-danger">Failed to load tokens: ${err.message}</p>`;
  }
}
