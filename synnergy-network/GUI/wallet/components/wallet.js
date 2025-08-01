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

document.getElementById('addressBtn').addEventListener('click', async () => {
    const wallet = JSON.parse(document.getElementById('walletJson').value);
    const account = parseInt(document.getElementById('account').value, 10) || 0;
    const index = parseInt(document.getElementById('index').value, 10) || 0;
    const res = await fetch('/api/wallet/address', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({Wallet: wallet, Account: account, Index: index})
    });
    const data = await res.json();
    document.getElementById('addressOutput').textContent = JSON.stringify(data, null, 2);
});

document.getElementById('signBtn').addEventListener('click', async () => {
    const wallet = JSON.parse(document.getElementById('walletJson').value);
    const tx = JSON.parse(document.getElementById('txJson').value);
    const account = parseInt(document.getElementById('account').value, 10) || 0;
    const index = parseInt(document.getElementById('index').value, 10) || 0;
    const res = await fetch('/api/wallet/sign', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({Wallet: wallet, Tx: tx, Account: account, Index: index, Gas: 0})
    });
    const data = await res.json();
    document.getElementById('signOutput').textContent = JSON.stringify(data, null, 2);
});

document.getElementById('opcodesBtn').addEventListener('click', async () => {
    const res = await fetch('/api/wallet/opcodes');
    const data = await res.json();
    document.getElementById('opcodesOutput').textContent = JSON.stringify(data, null, 2);
});

