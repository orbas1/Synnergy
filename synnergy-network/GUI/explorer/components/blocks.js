export async function loadBlocks() {
  const res = await fetch("/api/blocks");
  const blocks = await res.json();
  const tbody = document.querySelector("#blocks-table tbody");
  tbody.innerHTML = "";
  blocks
    .slice()
    .reverse()
    .forEach((b) => {
      const tr = document.createElement("tr");
      tr.innerHTML = `<td>${b.height}</td><td>${b.hash}</td><td>${b.txs}</td>`;
      tbody.appendChild(tr);
    });
}
