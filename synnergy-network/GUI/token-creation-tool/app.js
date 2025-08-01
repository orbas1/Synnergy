import { renderTokenForm } from './components/TokenForm.js';
import { renderTokenList } from './components/TokenList.js';

document.addEventListener('DOMContentLoaded', () => {
  const formContainer = document.getElementById('form-container');
  renderTokenForm(formContainer);
  const listContainer = document.getElementById('list-container');
  renderTokenList(listContainer);
});

async function createToken(data) {
  const res = await fetch('/api/tokens', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  });
  const result = document.getElementById('result');
  const out = await res.json();
  if (res.ok) {
    result.className = 'alert alert-success';
    result.textContent = `Token created with ID ${out.tokenId}`;
    renderTokenList(document.getElementById('list-container'));
  } else {
    result.className = 'alert alert-danger';
    result.textContent = out.error || 'Error creating token';
  }
  result.classList.remove('d-none');
}

document.addEventListener('submit', (e) => {
  if (e.target.id === 'tokenForm') {
    e.preventDefault();
    const form = e.target;
    const data = {
      name: form.name.value,
      symbol: form.symbol.value,
      decimals: parseInt(form.decimals.value, 10) || 0,
      standard: form.standard.value,
      fixedSupply: form.fixed.checked,
      supply: parseInt(form.supply.value, 10) || 0
    };
    createToken(data);
  }
});
