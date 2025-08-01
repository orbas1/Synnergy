import { listBridges, createBridge } from './services/api.js';
import { renderBridgeList } from './components/BridgeList.js';
import { attachBridgeForm } from './components/BridgeForm.js';

async function refresh() {
    const bridges = await listBridges();
    renderBridgeList(document.getElementById('bridgeListContainer'), bridges);
}

document.addEventListener('DOMContentLoaded', () => {
    attachBridgeForm(document.getElementById('bridgeFormContainer'), async (data) => {
        await createBridge(data);
        await refresh();
    });
    refresh();
});
