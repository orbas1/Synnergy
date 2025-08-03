import {
  listBridges,
  createBridge,
  authorizeRelayer,
  revokeRelayer,
  lockAndMint,
  burnAndRelease,
} from "./services/api.js";
import { renderBridgeList } from "./components/BridgeList.js";
import { attachBridgeForm } from "./components/BridgeForm.js";
import { attachRelayerForm } from "./components/RelayerForm.js";
import { attachLockMintForm } from "./components/LockMintForm.js";
import { attachBurnReleaseForm } from "./components/BurnReleaseForm.js";

async function refresh() {
  try {
    const bridges = await listBridges();
    renderBridgeList(
      document.getElementById("bridgeListContainer"),
      bridges
    );
  } catch (err) {
    console.error("Failed to load bridges", err);
  }
}

document.addEventListener("DOMContentLoaded", () => {
  attachBridgeForm(
    document.getElementById("bridgeFormContainer"),
    async (data) => {
      try {
        await createBridge(data);
        await refresh();
      } catch (err) {
        console.error("Bridge creation failed", err);
      }
    }
  );

  attachRelayerForm(document.getElementById("relayerFormContainer"), {
    onAuthorize: authorizeRelayer,
    onRevoke: revokeRelayer,
  });

  attachLockMintForm(
    document.getElementById("lockMintContainer"),
    lockAndMint
  );
  attachBurnReleaseForm(
    document.getElementById("burnReleaseContainer"),
    burnAndRelease
  );

  refresh();
});
