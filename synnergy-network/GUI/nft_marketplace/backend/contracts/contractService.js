const { ethers } = require('ethers');
const {
  providerUrl,
  privateKey,
  contractAddress,
} = require('../config');
const abi = require('./NFTMarketplaceABI.json');

const provider = new ethers.JsonRpcProvider(providerUrl);
const wallet = new ethers.Wallet(privateKey, provider);
const contract = new ethers.Contract(contractAddress, abi, wallet);

async function listNFT(tokenId, price) {
  return contract.list(tokenId, price);
}

async function buyNFT(id) {
  return contract.buy(id);
}

module.exports = { listNFT, buyNFT };
