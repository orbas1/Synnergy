async function fetchServices() {
  const res = await fetch("/api/services");
  if (!res.ok) {
    console.error("Failed to fetch services");
    return;
  }
  return res.json();
}

function renderServices(list) {
  const container = document.getElementById("services");
  container.innerHTML = "";
  list.forEach((svc) => {
    const card = document.createElement("div");
    card.className = "bg-white p-4 rounded shadow flex flex-col";
    card.innerHTML = `
            <h2 class="text-xl font-semibold mb-2">${svc.name}</h2>
            <p class="flex-grow">${svc.description}</p>
            <button class="mt-4 bg-blue-600 text-white px-4 py-2 rounded" data-id="${svc.id}">Buy</button>
        `;
    container.appendChild(card);
  });
}

async function init() {
  const services = await fetchServices();
  if (services) {
    renderServices(services);
  }

  document.getElementById("services").addEventListener("click", async (e) => {
    if (e.target.tagName === "BUTTON") {
      const id = e.target.getAttribute("data-id");
      const res = await fetch("/api/purchase", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id }),
      });
      if (res.ok) {
        alert("Purchase successful");
      } else {
        alert("Error purchasing");
      }
    }
  });
}

window.onload = init;
