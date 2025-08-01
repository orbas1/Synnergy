const service = require('../services/storageService');

exports.getDeals = async (req, res, next) => {
  try {
    const deals = await service.listDeals();
    res.json(deals);
  } catch (err) {
    next(err);
  }
};

exports.createDeal = async (req, res, next) => {
  try {
    const deal = await service.openDeal(req.body);
    res.status(201).json(deal);
  } catch (err) {
    next(err);
  }
};

exports.getDeal = async (req, res, next) => {
  try {
    const deal = await service.getDeal(req.params.id);
    if (!deal) return res.status(404).json({ error: 'Not found' });
    res.json(deal);
  } catch (err) {
    next(err);
  }
};
