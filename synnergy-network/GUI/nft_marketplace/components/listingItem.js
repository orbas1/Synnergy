export function createListingItem(listing) {
    const li = document.createElement('li');
    li.className = 'flex justify-between p-2 border-b border-gray-200';
    li.innerHTML = `
        <span>Token #${listing.tokenId} - Price: ${listing.price}</span>
        <button data-id="${listing.id}" class="bg-green-500 text-white px-2 py-1 rounded">Buy</button>
    `;
    return li;
}
