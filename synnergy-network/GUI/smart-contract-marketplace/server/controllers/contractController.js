const service = require('../services/contractService');

exports.list = async (req, res) => {
  const listings = await service.listContracts();
  res.json(listings);
};

exports.deploy = async (req, res) => {
  try {
    const { name, wasm } = req.body;
    const contract = await service.deployContract(name, wasm);
    res.status(201).json(contract);
  } catch (e) {
    res.status(500).json({ error: e.message });
  }
};

exports.get = async (req, res) => {
  const contract = await service.getContract(req.params.id);
  if (!contract) return res.status(404).end();
  res.json(contract);
};
