# Smart-Contract Marketplace GUI

This interface allows users to deploy and browse smart contracts on the Synnergy network.
It ships a lightweight Express backend that stores deployed contracts and
exposes a JSON API consumed by the HTML frontend.

## Getting Started

1. Install Node.js and run `npm install express morgan dotenv` inside the
   `server` directory.
2. Start the backend with `node server.js`.
3. Open `views/listings.html` in a browser to interact with the GUI.

The backend currently invokes placeholder commands to deploy contracts but can be
extended to call the Synnergy CLI.
