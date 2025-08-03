import path from "path";
import { fileURLToPath } from "url";
import fsPromises from "fs/promises";
import solc from "solc";
import { ethers } from "ethers";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const dbPath = path.join(__dirname, "..", "tokens.json");
const contractPath = path.join(
  __dirname,
  "..",
  "..",
  "..",
  "cmd",
  "smart_contracts",
  "token_factory.sol",
);

export async function createToken(data) {
  let tokens = [];
  try {
    const json = await fsPromises.readFile(dbPath, "utf8");
    tokens = JSON.parse(json);
  } catch {
    // If the file doesn't exist or is invalid, start with an empty array.
  }

  const tokenId = "0x" + Date.now().toString(16);
  tokens.push({ id: tokenId, ...data });
  await fsPromises.writeFile(dbPath, JSON.stringify(tokens, null, 2));
  return tokenId;
}

export async function listTokens() {
  try {
    const json = await fsPromises.readFile(dbPath, "utf8");
    return JSON.parse(json);
  } catch {
    return [];
  }
}

export async function compileContract() {
  const source = await fsPromises.readFile(contractPath, "utf8");
  const input = JSON.stringify({
    language: "Solidity",
    sources: { "token_factory.sol": { content: source } },
    settings: { outputSelection: { "*": { "*": ["abi", "evm.bytecode"] } } },
  });
  const output = JSON.parse(solc.compile(input));
  const contract = output.contracts["token_factory.sol"].TokenFactory;
  return { abi: contract.abi, bytecode: contract.evm.bytecode.object };
}

export async function deployContract(providerUrl, privateKey) {
  const { abi, bytecode } = await compileContract();
  const provider = new ethers.JsonRpcProvider(providerUrl);
  const wallet = new ethers.Wallet(privateKey, provider);
  const factory = new ethers.ContractFactory(abi, bytecode, wallet);
  const contract = await factory.deploy();
  await contract.waitForDeployment();
  return contract.target;
}
