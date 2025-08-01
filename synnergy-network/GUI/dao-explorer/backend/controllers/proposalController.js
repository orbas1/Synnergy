const service = require('../services/proposalService');

exports.list = (req, res) => {
  res.json(service.listProposals());
};

exports.get = (req, res) => {
  const proposal = service.getProposal(req.params.id);
  if (!proposal) return res.status(404).json({ message: 'Not found' });
  res.json(proposal);
};

exports.submit = (req, res) => {
  const proposal = service.submitProposal(req.body);
  res.status(201).json(proposal);
};

exports.vote = (req, res) => {
  const { id } = req.params;
  const { approve } = req.body;
  const proposal = service.castVote(id, approve);
  if (!proposal) return res.status(404).json({ message: 'Not found' });
  res.json(proposal);
};
