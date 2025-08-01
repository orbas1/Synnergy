document.getElementById('createBtn').addEventListener('click', async () => {
    const res = await fetch('/api/wallet/create');
    const data = await res.json();
    document.getElementById('createOutput').textContent = JSON.stringify(data, null, 2);
});

document.getElementById('importBtn').addEventListener('click', async () => {
    const mnemonic = document.getElementById('mnemonic').value;
    const res = await fetch('/api/wallet/import', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({mnemonic: mnemonic})
    });
    const data = await res.json();
    document.getElementById('importOutput').textContent = JSON.stringify(data, null, 2);
});
