const API = "/api/storage";

export async function renderPins() {
  const res = await fetch(`${API}/pins`);
  const data = await res.json();
  const tbody = document.querySelector("#pinsTable tbody");
  tbody.innerHTML = "";
  data.forEach((f) => {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td>${f.cid}</td>
      <td>${f.pinnedAt ? new Date(f.pinnedAt).toLocaleString() : ""}</td>
    `;
    tbody.appendChild(tr);
  });
}

export function initPinForm() {
  const form = document.getElementById("pinForm");
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    const data = Object.fromEntries(new FormData(form).entries());
    await fetch(`${API}/pin`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    });
    form.reset();
    renderPins();
  });
}

export async function initStorageSection() {
  await renderPins();
  initPinForm();
}
