const fs = require('fs');
const path = require('path');
const { exec } = require('child_process');

const DB_PATH = path.join(__dirname, 'contracts.json');

function load() {
  if (!fs.existsSync(DB_PATH)) return [];
  return JSON.parse(fs.readFileSync(DB_PATH));
}

function save(data) {
  fs.writeFileSync(DB_PATH, JSON.stringify(data, null, 2));
}

exports.listContracts = async () => {
  return load();
};

exports.deployContract = async (name, wasm) => {
  const contracts = load();
  const id = `c${Date.now()}`;
  const filename = path.join(__dirname, `${id}.wasm`);
  fs.writeFileSync(filename, Buffer.from(wasm, 'base64'));

  // Placeholder CLI deploy call
  await new Promise((resolve, reject) => {
    exec(`echo deploying ${filename}`, (err) => {
      if (err) reject(err); else resolve();
    });
  });

  const contract = { id, name };
  contracts.push(contract);
  save(contracts);
  return contract;
};

exports.getContract = async (id) => {
  const contracts = load();
  return contracts.find(c => c.id === id);
};
