const API = "/api/deals";

export async function renderDeals() {
  try {
    const res = await fetch(API);
    if (!res.ok) throw new Error(`Failed to fetch deals: ${res.status}`);
    const data = await res.json();
    const tbody = document.querySelector("#dealsTable tbody");
    tbody.innerHTML = "";
    data.forEach((d) => {
      const tr = document.createElement("tr");
      tr.innerHTML = `
        <td>${d.id}</td>
        <td>${d.listingId}</td>
        <td>${d.client}</td>
        <td>${d.duration}</td>
        <td>${new Date(d.createdAt).toLocaleString()}</td>
      `;
      tbody.appendChild(tr);
    });
  } catch (err) {
    console.error(err);
  }
}

export function initDealForm() {
  const form = document.getElementById("dealForm");
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    const data = Object.fromEntries(new FormData(form).entries());
    try {
      const res = await fetch(API, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      });
      if (!res.ok) throw new Error("Failed to open deal");
      form.reset();
      renderDeals();
    } catch (err) {
      console.error(err);
    }
  });
}
