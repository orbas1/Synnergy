const base = "/api";

export async function listBridges() {
  const r = await fetch(`${base}/bridges`);
  return r.json();
}

export async function createBridge(data) {
  await fetch(`${base}/bridges`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
}

export async function authorizeRelayer(data) {
  await fetch(`${base}/relayer/authorize`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
}

export async function revokeRelayer(data) {
  await fetch(`${base}/relayer/revoke`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
}

export async function lockAndMint(data) {
  await fetch(`${base}/lockmint`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
}

export async function burnAndRelease(data) {
  await fetch(`${base}/burnrelease`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
}
