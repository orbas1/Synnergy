async function loadNavbar() {
  const nav = document.getElementById("nav");
  if (!nav) return;
  try {
    const res = await fetch("components/navbar.html");
    if (!res.ok) throw new Error("navbar fetch failed");
    nav.innerHTML = await res.text();
  } catch (err) {
    console.error("Unable to load navbar", err);
  }
}

async function loadFooter() {
  const footer = document.getElementById("footer");
  if (!footer) return;
  try {
    const res = await fetch("components/footer.html");
    if (!res.ok) throw new Error("footer fetch failed");
    footer.innerHTML = await res.text();
  } catch (err) {
    console.error("Unable to load footer", err);
  }
}

async function loadListings() {
  const res = await fetch("../api/contracts").catch(() => ({ json: () => [] }));
  const listings = await res.json();
  const tbody = document.querySelector("#listings tbody");
  if (!tbody) return;
  tbody.innerHTML = listings
    .map(
      (c) =>
        `<tr><td><a href="detail.html?id=${c.id}">${c.id}</a></td><td>${c.name}</td></tr>`,
    )
    .join("");
}

async function loadDetail() {
  const params = new URLSearchParams(window.location.search);
  const id = params.get("id");
  if (!id) return;
  const res = await fetch(`../api/contracts/${id}`);
  const c = await res.json();
  document.getElementById("contractId").textContent = c.id;
  document.getElementById("contractName").textContent = c.name;
  document.getElementById("download").href = `../api/contracts/${id}/wasm`;
}

function bindDeploy() {
  const form = document.getElementById("deployForm");
  if (!form) return;
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    const name = document.getElementById("name").value;
    const wasm = document.getElementById("wasm").value;
    await fetch("../api/contracts", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, wasm }),
    });
    window.location.href = "listings.html";
  });
}

document.addEventListener("DOMContentLoaded", () => {
  loadNavbar();
  loadFooter();
  loadListings();
  loadDetail();
  bindDeploy();
});
