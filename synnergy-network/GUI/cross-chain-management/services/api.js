const base = "/api";

async function request(endpoint, options = {}) {
  const res = await fetch(`${base}${endpoint}`, {
    headers: { "Content-Type": "application/json", ...(options.headers || {}) },
    ...options,
  });
  if (!res.ok) {
    const message = await res.text();
    throw new Error(`Request failed with ${res.status}: ${message}`);
  }
  const contentType = res.headers.get("content-type") || "";
  if (contentType.includes("application/json")) {
    return res.json();
  }
  return res.text();
}

export const listBridges = () => request("/bridges");

export const createBridge = (data) =>
  request("/bridges", { method: "POST", body: JSON.stringify(data) });

export const authorizeRelayer = (data) =>
  request("/relayer/authorize", { method: "POST", body: JSON.stringify(data) });

export const revokeRelayer = (data) =>
  request("/relayer/revoke", { method: "POST", body: JSON.stringify(data) });

export const lockAndMint = (data) =>
  request("/lockmint", { method: "POST", body: JSON.stringify(data) });

export const burnAndRelease = (data) =>
  request("/burnrelease", { method: "POST", body: JSON.stringify(data) });

