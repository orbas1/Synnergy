import renderProposalList from './components/ProposalList.js';
import renderProposalDetail from './components/ProposalDetail.js';
import newProposalForm from './components/NewProposalForm.js';

const apiBase = '/backend/api';

async function fetchProposals() {
  const res = await fetch(`${apiBase}/proposals`);
  return res.json();
}

async function fetchProposal(id) {
  const res = await fetch(`${apiBase}/proposals/${id}`);
  return res.json();
}

async function submitProposal(data) {
  const res = await fetch(`${apiBase}/proposals`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  });
  return res.json();
}

async function vote(id, approve) {
  const res = await fetch(`${apiBase}/proposals/${id}/vote`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ approve })
  });
  return res.json();
}

function setupEventListeners() {
  document.getElementById('proposal-list').addEventListener('click', async e => {
    if (e.target.dataset.id) {
      const proposal = await fetchProposal(e.target.dataset.id);
      renderProposalDetail(proposal);
    }
  });

  document.body.addEventListener('submit', async e => {
    e.preventDefault();
    if (e.target.id === 'new-proposal-form') {
      const formData = new FormData(e.target);
      const proposal = await submitProposal(Object.fromEntries(formData));
      const proposals = await fetchProposals();
      renderProposalList(proposals);
      e.target.reset();
    } else if (e.target.id === 'vote-form') {
      const formData = new FormData(e.target);
      await vote(formData.get('id'), formData.get('approve') === 'true');
      const proposal = await fetchProposal(formData.get('id'));
      renderProposalDetail(proposal);
    }
  });
}

async function init() {
  document.getElementById('new-proposal').innerHTML = newProposalForm();
  const proposals = await fetchProposals();
  renderProposalList(proposals);
  setupEventListeners();
}

init();
