async function fetchListings() {
    const res = await fetch('/api/listings');
    const listings = await res.json();
    const container = document.getElementById('listingContainer');
    container.innerHTML = '';
    listings.forEach(l => {
        const li = document.createElement('li');
        li.innerHTML = `<span>Token #${l.tokenId} - Price: ${l.price}</span>` +
                       `<button data-id="${l.id}" class="bg-green-500 text-white px-2 py-1 rounded">Buy</button>`;
        container.appendChild(li);
    });
}

document.getElementById('listForm').addEventListener('submit', async e => {
    e.preventDefault();
    const tokenId = document.getElementById('tokenId').value;
    const price = document.getElementById('price').value;
    await fetch('/api/listings', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({ tokenId: Number(tokenId), price: Number(price) })
    });
    fetchListings();
});

document.getElementById('listingContainer').addEventListener('click', async e => {
    if (e.target.tagName === 'BUTTON') {
        const id = e.target.getAttribute('data-id');
        await fetch(`/api/listings/${id}/buy`, { method: 'POST' });
        fetchListings();
    }
});

window.onload = fetchListings;
