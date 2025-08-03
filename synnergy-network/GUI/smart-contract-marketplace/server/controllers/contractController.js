import * as service from "../services/contractService.js";

export async function list(req, res) {
  const listings = await service.listContracts();
  res.json(listings);
}

export async function deploy(req, res) {
  try {
    const { name, wasm } = req.body;
    const contract = await service.deployContract(name, wasm);
    res.status(201).json(contract);
  } catch (e) {
    res.status(500).json({ error: e.message });
  }
}

export async function get(req, res) {
  const contract = await service.getContract(req.params.id);
  if (!contract) return res.status(404).end();
  res.json(contract);
}

export async function remove(req, res) {
  await service.deleteContract(req.params.id);
  res.status(204).end();
}

export async function wasm(req, res) {
  const file = await service.getWasm(req.params.id);
  if (!file) return res.status(404).end();
  res.type("application/wasm").send(file);
}
