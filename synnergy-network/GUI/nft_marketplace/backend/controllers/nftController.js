const service = require("../services/marketplaceService");

/**
 * Create a new NFT listing.
 * Expects numeric `tokenId` and `price` in the request body.
 */
exports.create = async (req, res, next) => {
  try {
    const { tokenId, price } = req.body;
    if (typeof tokenId !== "number" || typeof price !== "number") {
      return res.status(400).json({ error: "tokenId and price must be numbers" });
    }
    const listing = await service.createListing(tokenId, price);
    res.status(201).json(listing);
  } catch (err) {
    next(err);
  }
};

/**
 * Return all active listings.
 */
exports.all = async (req, res, next) => {
  try {
    const listings = await service.listAll();
    res.json(listings);
  } catch (err) {
    next(err);
  }
};

/**
 * Purchase a listing by ID.
 */
exports.buy = async (req, res, next) => {
  try {
    await service.purchase(req.params.id);
    res.json({ success: true });
  } catch (err) {
    next(err);
  }
};
