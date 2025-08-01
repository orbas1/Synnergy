const { listNFT, buyNFT } = require('../contracts/contractService');

const listings = [];

module.exports = {
  async createListing(tokenId, price) {
    const listing = { id: listings.length, tokenId, price };
    listings.push(listing);
    await listNFT(tokenId, price);
    return listing;
  },

  async listAll() {
    return listings;
  },

  async purchase(id) {
    await buyNFT(id);
    const idx = listings.findIndex(l => l.id === Number(id));
    if (idx !== -1) listings.splice(idx, 1);
  }
};
