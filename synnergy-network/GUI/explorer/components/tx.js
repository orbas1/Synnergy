export async function searchTx(id) {
  const res = await fetch(`/api/tx/${id}`);
  if (!res.ok) return null;
  return await res.json();
}
