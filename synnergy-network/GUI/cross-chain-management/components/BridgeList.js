export function renderBridgeList(container, bridges) {
  container.innerHTML = "";

  const table = document.createElement("table");
  table.className = "table table-bordered";

  const thead = document.createElement("thead");
  const headerRow = document.createElement("tr");
  ["ID", "Source", "Target", "Relayer"].forEach((label) => {
    const th = document.createElement("th");
    th.textContent = label;
    headerRow.appendChild(th);
  });
  thead.appendChild(headerRow);
  table.appendChild(thead);

  const tbody = document.createElement("tbody");
  for (const b of bridges) {
    const row = document.createElement("tr");
    [b.id, b.source_chain, b.target_chain, b.relayer].forEach((val) => {
      const td = document.createElement("td");
      td.textContent = val;
      row.appendChild(td);
    });
    tbody.appendChild(row);
  }
  table.appendChild(tbody);

  if (!bridges.length) {
    const emptyRow = document.createElement("tr");
    const cell = document.createElement("td");
    cell.colSpan = 4;
    cell.textContent = "No bridges available";
    emptyRow.appendChild(cell);
    tbody.appendChild(emptyRow);
  }

  container.appendChild(table);
}
