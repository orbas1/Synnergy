// Placeholder service interacting with the NFTMarketplace smart contract
async function listNFT(tokenId, price) {
  console.log(`Listing token ${tokenId} for price ${price}`);
  // TODO: Integrate with CLI or blockchain RPC
}

async function buyNFT(id) {
  console.log(`Purchasing listing ${id}`);
  // TODO: Integrate with CLI or blockchain RPC
}

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
