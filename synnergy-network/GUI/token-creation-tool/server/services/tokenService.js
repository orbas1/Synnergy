import fs from 'fs';
import path from 'path';

const dbPath = path.join(path.dirname(new URL(import.meta.url).pathname), '..', 'tokens.json');

export async function createToken(data) {
  const tokens = fs.existsSync(dbPath) ? JSON.parse(fs.readFileSync(dbPath)) : [];
  const tokenId = '0x' + Date.now().toString(16);
  tokens.push({ id: tokenId, ...data });
  fs.writeFileSync(dbPath, JSON.stringify(tokens, null, 2));
  return tokenId;
}
