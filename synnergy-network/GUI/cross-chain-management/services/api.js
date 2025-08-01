const base = '/api';

export async function listBridges() {
    const r = await fetch(`${base}/bridges`);
    return r.json();
}

export async function createBridge(data) {
    await fetch(`${base}/bridges`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });
}
