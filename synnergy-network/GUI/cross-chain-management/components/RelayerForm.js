export function attachRelayerForm(
  container,
  { onAuthorize = async () => {}, onRevoke = async () => {} } = {},
) {
  container.innerHTML = `
        <form id="relayerForm" class="row gy-2 gx-2 align-items-center">
            <div class="col-auto">
                <input type="text" class="form-control" id="relayerAddr" placeholder="Relayer Address" required pattern="^0x[a-fA-F0-9]{40}$">
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
  const isValidAddress = (addr) => /^0x[a-fA-F0-9]{40}$/.test(addr);

  form.querySelector("#authBtn").addEventListener("click", async (e) => {
    e.preventDefault();
    const addr = addrInput.value.trim();
    if (!isValidAddress(addr)) {
      alert("Invalid address");
      return;
    }
    await onAuthorize({ addr });
    addrInput.value = "";
  });

  form.querySelector("#revokeBtn").addEventListener("click", async (e) => {
    e.preventDefault();
    const addr = addrInput.value.trim();
    if (!isValidAddress(addr)) {
      alert("Invalid address");
      return;
    }
    await onRevoke({ addr });
    addrInput.value = "";
  });
}
