export function createListingForm(onSubmit) {
    const form = document.createElement('form');
    form.id = 'listForm';
    form.className = 'space-y-4';
    form.innerHTML = `
        <div>
            <label class="block">Token ID</label>
            <input type="number" id="tokenId" class="w-full p-2 border rounded" />
        </div>
        <div>
            <label class="block">Price (SYNN)</label>
            <input type="number" id="price" class="w-full p-2 border rounded" />
        </div>
        <button type="submit" class="bg-blue-600 text-white px-4 py-2 rounded">List NFT</button>
    `;
    form.addEventListener('submit', onSubmit);
    return form;
}
