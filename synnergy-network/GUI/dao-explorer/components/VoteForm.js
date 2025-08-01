export default function voteForm(id) {
  return `
    <form id="vote-form" class="mt-2">
      <input type="hidden" name="id" value="${id}">
      <button name="approve" value="true" class="bg-green-500 text-white px-2 py-1 mr-2 rounded">Approve</button>
      <button name="approve" value="false" class="bg-red-500 text-white px-2 py-1 rounded">Reject</button>
    </form>
  `;
}
