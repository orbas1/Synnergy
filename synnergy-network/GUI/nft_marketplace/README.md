# NFT Marketplace GUI

This interface allows users to list and purchase NFTs on the Synnergy Network.
It is composed of a small Node.js backend and a modular frontend built with
Tailwind CSS.

## Structure

- `index.html` – entry page loading ES modules
- `components/` – reusable UI components
- `views/` – HTML views (served statically)
- `style.css` – additional global styling
- `app.js` – bootstrap code assembling components
- `backend/` – Express service interacting with the smart contract
- `smart_contracts/` – Solidity code compiled/deployed separately

Run `npm install` inside `backend/` and then `node server.js` to start the API.
The frontend can be opened directly in the browser when the backend is running.
