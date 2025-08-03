// Placeholder service interacting with the NFTMarketplace smart contract
async function listNFT(tokenId, price) {
  console.log(`Listing token ${tokenId} for price ${price}`);
  // TODO: Integrate with CLI or blockchain RPC
}

async function buyNFT(id) {
  console.log(`Purchasing listing ${id}`);
  // TODO: Integrate with CLI or blockchain RPC
}

// In-memory listing store. Replace with persistent storage in production.
const listings = [];

module.exports = {
  /**
   * Create a new listing and return it.
   */
  async createListing(tokenId, price) {
    if (typeof tokenId !== "number" || typeof price !== "number") {
      throw new Error("tokenId and price must be numbers");
    }
    const listing = { id: listings.length, tokenId, price };
    listings.push(listing);
    await listNFT(tokenId, price);
    return listing;
  },

  /**
   * Return all current listings.
   */
  async listAll() {
    return listings;
  },

  /**
   * Purchase a listing by ID, removing it from the store.
   */
  async purchase(id) {
    const idx = listings.findIndex((l) => l.id === Number(id));
    if (idx === -1) {
      throw new Error("Listing not found");
    }
    await buyNFT(id);
    listings.splice(idx, 1);
  },
};
