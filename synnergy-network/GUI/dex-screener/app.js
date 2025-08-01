async function loadPools(onSelect) {
    const res = await fetch('/api/pools');
    const pools = await res.json();
    const tbody = document.getElementById('pools');
    tbody.innerHTML = '';
    for (const p of pools) {
        const tr = document.createElement('tr');
        tr.classList.add('cursor-pointer','hover:bg-gray-100');
        tr.dataset.pair = `${p.token_a}/${p.token_b}`;
        tr.innerHTML = `
            <td>${p.id}</td>
            <td>${p.token_a}/${p.token_b}</td>
            <td>${p.res_a}</td>
            <td>${p.res_b}</td>
            <td>${p.fee_bps}</td>
        `;
        tr.addEventListener('click', () => onSelect(p));
        tbody.appendChild(tr);
    }
}

function initChart() {
    const chartEl = document.getElementById('chart');
    const chart = LightweightCharts.createChart(chartEl, { width: chartEl.clientWidth, height: chartEl.clientHeight });
    const lineSeries = chart.addLineSeries();
    return { chart, lineSeries };
}

function generateSeries(p) {
    const data = [];
    let price = p.res_a / p.res_b || 1;
    for (let i = 29; i >= 0; i--) {
        const t = Date.now() / 1000 - i * 60;
        price *= 1 + (Math.random() - 0.5) / 20;
        data.push({ time: Math.round(t), value: price });
    }
    return data;
}

function init() {
    const { lineSeries } = initChart();
    loadPools(pool => {
        lineSeries.setData(generateSeries(pool));
    });
}

window.onload = init;
