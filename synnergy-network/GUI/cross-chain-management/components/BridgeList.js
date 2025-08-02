export function renderBridgeList(container, bridges) {
  let html = `<table class="table table-bordered"><thead><tr><th>ID</th><th>Source</th><th>Target</th><th>Relayer</th></tr></thead><tbody>`;
  for (const b of bridges) {
    html += `<tr><td>${b.id}</td><td>${b.source_chain}</td><td>${b.target_chain}</td><td>${b.relayer}</td></tr>`;
  }
  html += "</tbody></table>";
  container.innerHTML = html;
}
