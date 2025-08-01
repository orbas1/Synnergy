export function createHeader() {
    const nav = document.createElement('nav');
    nav.className = 'bg-blue-800 p-4 text-white flex items-center';
    nav.innerHTML = `
        <img src="assets/logo.png" alt="Synnergy" class="h-8 mr-2" />
        <h1 class="text-2xl font-bold">Synnergy NFT Marketplace</h1>
    `;
    return nav;
}
