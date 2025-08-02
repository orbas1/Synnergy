export function attachLockMintForm(container, onSubmit) {
  container.innerHTML = `
        <form id="lockMintForm" class="row g-2">
            <div class="col-md-3">
                <input type="number" class="form-control" id="assetId" placeholder="Asset ID" required>
            </div>
            <div class="col-md-3">
                <input type="number" class="form-control" id="amount" placeholder="Amount" required>
            </div>
            <div class="col-md-4">
                <input type="text" class="form-control" id="proof" placeholder="SPV Proof (hex)" required>
            </div>
            <div class="col-md-2">
                <button class="btn btn-primary w-100" type="submit">Lock & Mint</button>
            </div>
        </form>`;
  const form = container.querySelector("#lockMintForm");
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    await onSubmit({
      asset_id: parseInt(form.assetId.value),
      amount: parseInt(form.amount.value),
      proof: form.proof.value,
    });
    form.reset();
  });
}
