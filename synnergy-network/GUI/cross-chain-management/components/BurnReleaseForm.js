export function attachBurnReleaseForm(container, onSubmit) {
  container.innerHTML = `
        <form id="burnReleaseForm" class="row g-2">
            <div class="col-md-3">
                <input type="number" class="form-control" id="assetId" placeholder="Asset ID" required>
            </div>
            <div class="col-md-4">
                <input type="text" class="form-control" id="recipient" placeholder="Recipient" required>
            </div>
            <div class="col-md-3">
                <input type="number" class="form-control" id="amount" placeholder="Amount" required>
            </div>
            <div class="col-md-2">
                <button class="btn btn-warning w-100" type="submit">Burn & Release</button>
            </div>
        </form>`;
  const form = container.querySelector("#burnReleaseForm");
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    await onSubmit({
      asset_id: parseInt(form.assetId.value),
      to: form.recipient.value,
      amount: parseInt(form.amount.value),
    });
    form.reset();
  });
}
