async function fetchListings() {
  try {
    const res = await fetch("/api/listings");
    if (!res.ok) throw new Error("listings request failed");
    const listings = await res.json();
    const container = document.getElementById("listingContainer");
    if (!container) return;
    container.innerHTML = "";
    listings.forEach((l) => {
      const li = document.createElement("li");
      li.innerHTML =
        `<span>Token #${l.tokenId} - Price: ${l.price}</span>` +
        `<button data-id="${l.id}" class="bg-green-500 text-white px-2 py-1 rounded">Buy</button>`;
      container.appendChild(li);
    });
  } catch (err) {
    console.error("Failed to load listings", err);
  }
}

document.getElementById("listForm")?.addEventListener("submit", async (e) => {
  e.preventDefault();
  const tokenId = document.getElementById("tokenId")?.value;
  const price = document.getElementById("price")?.value;
  try {
    await fetch("/api/listings", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tokenId: Number(tokenId), price: Number(price) }),
    });
    fetchListings();
  } catch (err) {
    console.error("Failed to list NFT", err);
  }
});

document.getElementById("listingContainer")?.addEventListener(
  "click",
  async (e) => {
    if (e.target.tagName === "BUTTON") {
      const id = e.target.getAttribute("data-id");
      try {
        await fetch(`/api/listings/${id}/buy`, { method: "POST" });
        fetchListings();
      } catch (err) {
        console.error("Purchase failed", err);
      }
    }
  },
);

window.addEventListener("load", fetchListings);
