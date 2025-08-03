async function safeFetch(url, options = {}) {
  const res = await fetch(url, options);
  if (!res.ok) {
    throw new Error(`request failed: ${res.status}`);
  }
  return res.json();
}

function writeOutput(id, content) {
  document.getElementById(id).textContent = content;
}

document.getElementById("createBtn").addEventListener("click", async () => {
  try {
    const data = await safeFetch("/api/wallet/create");
    writeOutput("createOutput", JSON.stringify(data, null, 2));
  } catch (err) {
    writeOutput("createOutput", `Error: ${err.message}`);
  }
});

document.getElementById("importBtn").addEventListener("click", async () => {
  try {
    const mnemonic = document.getElementById("mnemonic").value;
    const data = await safeFetch("/api/wallet/import", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ mnemonic }),
    });
    writeOutput("importOutput", JSON.stringify(data, null, 2));
  } catch (err) {
    writeOutput("importOutput", `Error: ${err.message}`);
  }
});

document.getElementById("addressBtn").addEventListener("click", async () => {
  try {
    const wallet = JSON.parse(document.getElementById("walletJson").value);
    const account = parseInt(document.getElementById("account").value, 10) || 0;
    const index = parseInt(document.getElementById("index").value, 10) || 0;
    const data = await safeFetch("/api/wallet/address", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ Wallet: wallet, Account: account, Index: index }),
    });
    writeOutput("addressOutput", JSON.stringify(data, null, 2));
  } catch (err) {
    writeOutput("addressOutput", `Error: ${err.message}`);
  }
});

document.getElementById("signBtn").addEventListener("click", async () => {
  try {
    const wallet = JSON.parse(document.getElementById("walletJson").value);
    const tx = JSON.parse(document.getElementById("txJson").value);
    const account = parseInt(document.getElementById("account").value, 10) || 0;
    const index = parseInt(document.getElementById("index").value, 10) || 0;
    const data = await safeFetch("/api/wallet/sign", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ Wallet: wallet, Tx: tx, Account: account, Index: index, Gas: 0 }),
    });
    writeOutput("signOutput", JSON.stringify(data, null, 2));
  } catch (err) {
    writeOutput("signOutput", `Error: ${err.message}`);
  }
});

document.getElementById("opcodesBtn").addEventListener("click", async () => {
  try {
    const data = await safeFetch("/api/wallet/opcodes");
    writeOutput("opcodesOutput", JSON.stringify(data, null, 2));
  } catch (err) {
    writeOutput("opcodesOutput", `Error: ${err.message}`);
  }
});
