import {
  createToken,
  listTokens,
  deployContract,
} from "../services/tokenService.js";

export async function create(req, res) {
  try {
    const tokenId = await createToken(req.body);
    res.json({ tokenId });
  } catch (err) {
    res.status(400).json({ error: err.message });
  }
}

export async function index(_req, res) {
  try {
    const tokens = await listTokens();
    res.json(tokens);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
}

export async function deploy(req, res) {
  try {
    const { rpcUrl, privateKey } = req.body;
    const address = await deployContract(rpcUrl, privateKey);
    res.json({ address });
  } catch (err) {
    res.status(400).json({ error: err.message });
  }
}
