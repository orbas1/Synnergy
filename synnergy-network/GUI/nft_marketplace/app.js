async function fetchListings() {
  const container = document.getElementById("listingContainer");
  try {
    const res = await fetch("/api/listings");
    if (!res.ok) throw new Error("Failed to load listings");
    const listings = await res.json();
    container.innerHTML = "";
    listings.forEach((l) => {
      const li = document.createElement("li");
      li.innerHTML =
        `<span>Token #${l.tokenId} - Price: ${l.price}</span>` +
        `<button data-id="${l.id}" class="bg-green-500 text-white px-2 py-1 rounded">Buy</button>`;
      container.appendChild(li);
    });
  } catch (err) {
    console.error(err);
    container.innerHTML = "<li class=\"text-red-500\">Error loading listings</li>";
  }
}

document.getElementById("listForm").addEventListener("submit", async (e) => {
  e.preventDefault();
  const tokenId = Number(document.getElementById("tokenId").value);
  const price = Number(document.getElementById("price").value);
  try {
    const res = await fetch("/api/listings", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tokenId, price }),
    });
    if (!res.ok) throw new Error("Failed to create listing");
    fetchListings();
  } catch (err) {
    console.error(err);
  }
});

document
  .getElementById("listingContainer")
  .addEventListener("click", async (e) => {
    if (e.target.tagName === "BUTTON") {
      const id = e.target.getAttribute("data-id");
      try {
        const res = await fetch(`/api/listings/${id}/buy`, { method: "POST" });
        if (!res.ok) throw new Error("Failed to purchase listing");
        fetchListings();
      } catch (err) {
        console.error(err);
      }
    }
  });

window.onload = fetchListings;
