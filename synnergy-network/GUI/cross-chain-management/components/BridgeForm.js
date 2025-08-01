export function attachBridgeForm(container, onSubmit) {
    container.innerHTML = `
        <form id="bridgeForm" class="row g-3">
            <div class="col-md-4">
                <input type="text" class="form-control" id="sourceChain" placeholder="Source Chain" required>
            </div>
            <div class="col-md-4">
                <input type="text" class="form-control" id="targetChain" placeholder="Target Chain" required>
            </div>
            <div class="col-md-4">
                <input type="text" class="form-control" id="relayer" placeholder="Relayer Address" required>
            </div>
            <div class="col-12">
                <button class="btn btn-primary" type="submit">Register Bridge</button>
            </div>
        </form>`;
    const form = container.querySelector('#bridgeForm');
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const data = {
            source_chain: form.sourceChain.value,
            target_chain: form.targetChain.value,
            relayer: form.relayer.value
        };
        await onSubmit(data);
        form.reset();
    });
}
