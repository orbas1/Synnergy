const fs = require('fs');
const path = require('path');
const { dataPath } = require('../config');

function load() {
  try {
    return JSON.parse(fs.readFileSync(dataPath, 'utf8'));
  } catch {
    return { listings: [], deals: [], files: [], storages: [] };
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

exports.getListing = async (id) => {
  const db = load();
  return db.listings.find(l => l.id === id);
};

exports.getDeal = async (id) => {
  const db = load();
  return db.deals.find(d => d.id === id);
};

exports.pin = async (cid, meta) => {
  const db = load();
  const file = { cid, meta, pinnedAt: new Date().toISOString() };
  db.files.push(file);
  save(db);
  return file;
};

exports.listPins = async () => {
  const db = load();
  return db.files;
};

exports.retrieve = async (cid) => {
  const db = load();
  return db.files.find(f => f.cid === cid);
};

exports.exists = async (cid) => {
  const db = load();
  return db.files.some(f => f.cid === cid);
};

exports.createStorage = async (input) => {
  const db = load();
  const storage = {
    id: Date.now().toString(),
    owner: input.owner,
    capacityGB: Number(input.capacityGB),
    createdAt: new Date().toISOString()
  };
  db.storages.push(storage);
  save(db);
  return storage;
};

exports.listStorages = async () => {
  const db = load();
  return db.storages;
};

