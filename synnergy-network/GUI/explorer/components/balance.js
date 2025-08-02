export async function checkBalance(addr) {
  const res = await fetch(`/api/balance/${addr}`);
  if (!res.ok) return null;
  const data = await res.json();
  return data.balance;
}
