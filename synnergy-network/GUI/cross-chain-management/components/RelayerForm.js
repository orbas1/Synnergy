export function attachRelayerForm(container, { onAuthorize, onRevoke }) {
  container.innerHTML = `
        <form id="relayerForm" class="row gy-2 gx-2 align-items-center">
            <div class="col-auto">
                <input type="text" class="form-control" id="relayerAddr" placeholder="Relayer Address" required>
            </div>
            <div class="col-auto">
                <button class="btn btn-success" id="authBtn">Authorize</button>
            </div>
            <div class="col-auto">
                <button class="btn btn-danger" id="revokeBtn">Revoke</button>
            </div>
        </form>`;

  const form = container.querySelector("#relayerForm");
  const addrInput = form.querySelector("#relayerAddr");
  form.querySelector("#authBtn").addEventListener("click", async (e) => {
    e.preventDefault();
    await onAuthorize({ addr: addrInput.value });
    addrInput.value = "";
  });
  form.querySelector("#revokeBtn").addEventListener("click", async (e) => {
    e.preventDefault();
    await onRevoke({ addr: addrInput.value });
    addrInput.value = "";
  });
}
