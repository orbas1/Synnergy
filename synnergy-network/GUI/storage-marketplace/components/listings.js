const API = "/api/listings";

export async function renderListings() {
  try {
    const res = await fetch(API);
    if (!res.ok) throw new Error(`Failed to fetch listings: ${res.status}`);
    const data = await res.json();
    const tbody = document.querySelector("#listingsTable tbody");
    tbody.innerHTML = "";
    data.forEach((l) => {
      const tr = document.createElement("tr");
      tr.innerHTML = `
        <td>${l.id}</td>
        <td>${l.provider}</td>
        <td>${l.pricePerGB}</td>
        <td>${l.capacityGB}</td>
        <td>${new Date(l.createdAt).toLocaleString()}</td>
      `;
      tbody.appendChild(tr);
    });
  } catch (err) {
    console.error(err);
  }
}

export function initListingForm() {
  const form = document.getElementById("listingForm");
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    const data = Object.fromEntries(new FormData(form).entries());
    try {
      const res = await fetch(API, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      });
      if (!res.ok) throw new Error("Failed to create listing");
      form.reset();
      renderListings();
    } catch (err) {
      console.error(err);
    }
  });
}
