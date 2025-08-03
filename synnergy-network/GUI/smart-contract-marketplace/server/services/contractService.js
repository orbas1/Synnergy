import fs from "fs";
import path from "path";
import { exec } from "child_process";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const DB_PATH =
  process.env.DB_FILE || path.join(__dirname, "../data/contracts.json");
const CLI = process.env.CLI_PATH || "synnergy";

if (!fs.existsSync(path.dirname(DB_PATH))) {
  fs.mkdirSync(path.dirname(DB_PATH), { recursive: true });
}

function load() {
  if (!fs.existsSync(DB_PATH)) return [];
  return JSON.parse(fs.readFileSync(DB_PATH));
}

function save(data) {
  fs.writeFileSync(DB_PATH, JSON.stringify(data, null, 2));
}

export async function listContracts() {
  return load();
}

export async function deployContract(name, wasm) {
  const contracts = load();
  const id = `c${Date.now()}`;
  const filename = path.join(__dirname, `${id}.wasm`);
  fs.writeFileSync(filename, Buffer.from(wasm, "base64"));

  // Deploy contract via Synnergy CLI
  await new Promise((resolve, reject) => {
    exec(`${CLI} contracts deploy ${filename}`, (err) => {
      if (err) reject(err);
      else resolve();
    });
  });

  const contract = { id, name };
  contracts.push(contract);
  save(contracts);
  return contract;
}

export async function getContract(id) {
  const contracts = load();
  return contracts.find((c) => c.id === id);
}

export async function deleteContract(id) {
  let contracts = load();
  contracts = contracts.filter((c) => c.id !== id);
  save(contracts);
}

export async function getWasm(id) {
  const file = path.join(__dirname, `${id}.wasm`);
  if (!fs.existsSync(file)) return null;
  return fs.readFileSync(file);
}
