const fs = require('fs');
const path = require('path');

const dataPath = path.join(__dirname, 'data.json');

function load() {
  try {
    return JSON.parse(fs.readFileSync(dataPath, 'utf8'));
  } catch {
    return { listings: [], deals: [] };
  }
}

function save(db) {
  fs.writeFileSync(dataPath, JSON.stringify(db, null, 2));
}

exports.listListings = async () => {
  const db = load();
  return db.listings;
};

exports.createListing = async (input) => {
  const db = load();
  const listing = {
    id: Date.now().toString(),
    provider: input.provider,
    pricePerGB: Number(input.pricePerGB),
    capacityGB: Number(input.capacityGB),
    createdAt: new Date().toISOString()
  };
  db.listings.push(listing);
  save(db);
  return listing;
};

exports.listDeals = async () => {
  const db = load();
  return db.deals;
};

exports.openDeal = async (input) => {
  const db = load();
  const deal = {
    id: Date.now().toString(),
    listingId: input.listingId,
    client: input.client,
    duration: input.duration,
    createdAt: new Date().toISOString()
  };
  db.deals.push(deal);
  save(db);
  return deal;
};
