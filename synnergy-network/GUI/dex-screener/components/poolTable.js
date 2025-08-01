export async function renderPools() {
    const res = await fetch('/api/pools');
    const pools = await res.json();
    const tbody = document.getElementById('pools');
    tbody.innerHTML = '';
    for (const p of pools) {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${p.id}</td>
            <td>${p.token_a}/${p.token_b}</td>
            <td>${p.res_a}</td>
            <td>${p.res_b}</td>
            <td>${p.fee_bps}</td>
        `;
        tbody.appendChild(tr);
    }
}
