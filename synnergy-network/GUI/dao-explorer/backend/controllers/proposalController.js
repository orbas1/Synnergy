const service = require('../services/proposalService');

exports.list = async (req, res, next) => {
  try {
    res.json(await service.listProposals());
  } catch (err) {
    next(err);
  }
};

exports.get = async (req, res, next) => {
  try {
    const proposal = await service.getProposal(req.params.id);
    if (!proposal) return res.status(404).json({ message: 'Not found' });
    res.json(proposal);
  } catch (err) {
    next(err);
  }
};

exports.submit = async (req, res, next) => {
  try {
    const proposal = await service.submitProposal(req.body);
    res.status(201).json(proposal);
  } catch (err) {
    next(err);
  }
};

exports.vote = async (req, res, next) => {
  try {
    const { id } = req.params;
    const { approve } = req.body;
    const proposal = await service.castVote(id, approve);
    if (!proposal) return res.status(404).json({ message: 'Not found' });
    res.json(proposal);
  } catch (err) {
    next(err);
  }
};

exports.execute = async (req, res, next) => {
  try {
    const tx = await service.execute(req.params.id);
    res.json({ tx });
  } catch (err) {
    next(err);
  }
};

exports.balance = async (req, res, next) => {
  try {
    const bal = await service.balance(req.params.id, req.params.asset);
    res.json({ balance: bal });
  } catch (err) {
    next(err);
  }
};
