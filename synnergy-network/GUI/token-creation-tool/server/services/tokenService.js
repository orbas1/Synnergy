import fs from "fs";
import path from "path";
import solc from "solc";
import { ethers } from "ethers";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

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
  const tokens = fs.existsSync(dbPath)
    ? JSON.parse(fs.readFileSync(dbPath))
    : [];
  const tokenId = "0x" + Date.now().toString(16);
  tokens.push({ id: tokenId, ...data });
  fs.writeFileSync(dbPath, JSON.stringify(tokens, null, 2));
  return tokenId;
}

export async function listTokens() {
  return fs.existsSync(dbPath) ? JSON.parse(fs.readFileSync(dbPath)) : [];
}

export async function compileContract() {
  const source = fs.readFileSync(contractPath, "utf8");
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
