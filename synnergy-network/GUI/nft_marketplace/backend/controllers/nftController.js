const service = require('../services/marketplaceService');

exports.create = async (req, res) => {
  const { tokenId, price } = req.body;
  const listing = await service.createListing(tokenId, price);
  res.json(listing);
};

exports.all = async (req, res) => {
  const listings = await service.listAll();
  res.json(listings);
};

exports.buy = async (req, res) => {
  await service.purchase(req.params.id);
  res.json({ success: true });
};
