import { listBridges, createBridge, authorizeRelayer, revokeRelayer, lockAndMint, burnAndRelease } from './services/api.js';
import { renderBridgeList } from './components/BridgeList.js';
import { attachBridgeForm } from './components/BridgeForm.js';
import { attachRelayerForm } from './components/RelayerForm.js';
import { attachLockMintForm } from './components/LockMintForm.js';
import { attachBurnReleaseForm } from './components/BurnReleaseForm.js';

async function refresh() {
    const bridges = await listBridges();
    renderBridgeList(document.getElementById('bridgeListContainer'), bridges);
}

document.addEventListener('DOMContentLoaded', () => {
    attachBridgeForm(document.getElementById('bridgeFormContainer'), async (data) => {
        await createBridge(data);
        await refresh();
    });

    attachRelayerForm(document.getElementById('relayerFormContainer'), {
        onAuthorize: authorizeRelayer,
        onRevoke: revokeRelayer
    });

    attachLockMintForm(document.getElementById('lockMintContainer'), lockAndMint);
    attachBurnReleaseForm(document.getElementById('burnReleaseContainer'), burnAndRelease);

    refresh();
});
