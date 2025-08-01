const service = require('../services/storageService');

exports.getListings = async (req, res, next) => {
  try {
    const listings = await service.listListings();
    res.json(listings);
  } catch (err) {
    next(err);
  }
};

exports.createListing = async (req, res, next) => {
  try {
    const listing = await service.createListing(req.body);
    res.status(201).json(listing);
  } catch (err) {
    next(err);
  }
};

exports.getListing = async (req, res, next) => {
  try {
    const listing = await service.getListing(req.params.id);
    if (!listing) return res.status(404).json({ error: 'Not found' });
    res.json(listing);
  } catch (err) {
    next(err);
  }
};

