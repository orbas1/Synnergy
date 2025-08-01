export default function newProposalForm() {
  return `
    <h2 class="text-xl font-bold">New Proposal</h2>
    <form id="new-proposal-form" class="flex flex-col gap-2 mt-2">
      <input name="title" class="border p-2" placeholder="Title" required>
      <textarea name="description" class="border p-2" placeholder="Description" required></textarea>
      <button class="bg-blue-600 text-white px-4 py-2 rounded" type="submit">Submit</button>
    </form>
  `;
}
