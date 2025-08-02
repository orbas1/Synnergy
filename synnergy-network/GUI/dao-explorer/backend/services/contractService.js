const { ethers } = require("ethers");
const config = require("../config/config");
const abi = require("../../../../cmd/smart_contracts/dao_explorer.json");

// Connect to network (assumes local JSON-RPC node)
const provider = new ethers.JsonRpcProvider(config.rpcUrl);
const signer = provider.getSigner();
const contract = new ethers.Contract(config.contractAddress, abi, signer);

async function submitProposal(data) {
  const tx = await contract.submit(ethers.toUtf8Bytes(JSON.stringify(data)));
  await tx.wait();
  return tx.hash;
}

async function vote(proposalId, approve) {
  const tx = await contract.vote(ethers.hexZeroPad(proposalId, 32), approve);
  await tx.wait();
  return tx.hash;
}

async function getProposal(id) {
  const result = await contract.get(ethers.hexZeroPad(id, 32));
  return JSON.parse(ethers.toUtf8String(result));
}

async function listProposals() {
  const result = await contract.list();
  return JSON.parse(ethers.toUtf8String(result));
}

async function executeProposal(id) {
  const tx = await contract.execute(ethers.hexZeroPad(id, 32));
  await tx.wait();
  return tx.hash;
}

async function balanceOfAsset(asset) {
  return await contract.balanceOf(asset);
}

module.exports = {
  submitProposal,
  vote,
  getProposal,
  listProposals,
  executeProposal,
  balanceOfAsset,
};
