export function renderTokenForm(container) {
  container.innerHTML = `
    <form id="tokenForm" class="row g-3 needs-validation" novalidate>
      <div class="col-md-6">
        <label class="form-label">Name</label>
        <input type="text" name="name" class="form-control" required />
      </div>
      <div class="col-md-6">
        <label class="form-label">Symbol</label>
        <input type="text" name="symbol" class="form-control" required />
      </div>
      <div class="col-md-4">
        <label class="form-label">Decimals</label>
        <input type="number" name="decimals" class="form-control" value="18" min="0" />
      </div>
      <div class="col-md-4">
        <label class="form-label">Standard</label>
        <input type="text" name="standard" class="form-control" placeholder="StdSYN20" />
      </div>
      <div class="col-md-4">
        <label class="form-label">Initial Supply</label>
        <input type="number" name="supply" class="form-control" value="0" min="0" />
      </div>
      <div class="col-12 form-check">
        <input class="form-check-input" type="checkbox" name="fixed" id="fixed" />
        <label class="form-check-label" for="fixed">Fixed Supply</label>
      </div>
      <div class="col-12">
        <button class="btn btn-primary" type="submit">Create Token</button>
      </div>
    </form>
  `;
}
