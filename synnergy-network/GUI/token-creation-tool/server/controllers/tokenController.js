import { createToken } from '../services/tokenService.js';

export async function create(req, res) {
  try {
    const tokenId = await createToken(req.body);
    res.json({ tokenId });
  } catch (err) {
    res.status(400).json({ error: err.message });
  }
}
