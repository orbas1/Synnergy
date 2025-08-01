import { loadBlocks } from './components/blocks.js';
import { searchTx } from './components/tx.js';
import { checkBalance } from './components/balance.js';

async function showInfo() {
    const res = await fetch('/api/info');
    const info = await res.json();
    document.getElementById('info').textContent = `Height ${info.height} - ${info.hash}`;
}

window.addEventListener('load', () => {
    loadBlocks();
    showInfo();

    document.getElementById('tx-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const id = e.target.txid.value.trim();
        const tx = await searchTx(id);
        document.getElementById('tx-result').textContent = tx ? JSON.stringify(tx, null, 2) : 'Not found';
    });

    document.getElementById('bal-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const addr = e.target.address.value.trim();
        const bal = await checkBalance(addr);
        document.getElementById('bal-result').textContent = bal !== null ? bal : 'Not found';
    });
});

