const API = "/api/storage";

export async function renderPins() {
  try {
    const res = await fetch(`${API}/pins`);
    if (!res.ok) throw new Error(`Failed to fetch pins: ${res.status}`);
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
  } catch (err) {
    console.error(err);
  }
}

export function initPinForm() {
  const form = document.getElementById("pinForm");
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    const data = Object.fromEntries(new FormData(form).entries());
    try {
      const res = await fetch(`${API}/pin`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      });
      if (!res.ok) throw new Error("Failed to pin file");
      form.reset();
      renderPins();
    } catch (err) {
      console.error(err);
    }
  });
}

export async function initStorageSection() {
  await renderPins();
  initPinForm();
}
