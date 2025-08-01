import { createHeader } from './components/header.js';
import { createListingForm } from './components/listingForm.js';
import { createListingItem } from './components/listingItem.js';

async function fetchListings(listingContainer) {
    const res = await fetch('/api/listings');
    const listings = await res.json();
    listingContainer.innerHTML = '';
    listings.forEach(l => {
        listingContainer.appendChild(createListingItem(l));
    });
}

function init() {
    const app = document.getElementById('app');
    app.appendChild(createHeader());

    const container = document.createElement('div');
    container.className = 'container mx-auto p-4';

    const formWrapper = document.createElement('div');
    formWrapper.className = 'bg-white shadow-md rounded p-4 mb-6';
    const listingContainer = document.createElement('ul');
    listingContainer.id = 'listingContainer';
    listingContainer.className = 'space-y-4';

    const form = createListingForm(async e => {
        e.preventDefault();
        const tokenId = document.getElementById('tokenId').value;
        const price = document.getElementById('price').value;
        await fetch('/api/listings', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({ tokenId: Number(tokenId), price: Number(price) })
        });
        await fetchListings(listingContainer);
        e.target.reset();
    });

    formWrapper.innerHTML = '<h2 class="text-xl font-semibold mb-4">List an NFT</h2>';
    formWrapper.appendChild(form);

    const listWrapper = document.createElement('div');
    listWrapper.className = 'bg-white shadow-md rounded p-4';
    listWrapper.innerHTML = '<h2 class="text-xl font-semibold mb-4">Marketplace Listings</h2>';
    listWrapper.appendChild(listingContainer);

    container.appendChild(formWrapper);
    container.appendChild(listWrapper);
    app.appendChild(container);

    listingContainer.addEventListener('click', async e => {
        if (e.target.tagName === 'BUTTON') {
            const id = e.target.getAttribute('data-id');
            await fetch(`/api/listings/${id}/buy`, { method: 'POST' });
            await fetchListings(listingContainer);
        }
    });

    fetchListings(listingContainer);
}

window.addEventListener('DOMContentLoaded', init);
