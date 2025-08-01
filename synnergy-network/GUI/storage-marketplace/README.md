# Storage Marketplace GUI

This interface allows users to list storage offerings, open deals and pin files
on Synnergy Network. A simple Express backend exposes REST APIs that proxy calls
to the blockchain. The frontend uses vanilla ES modules with Bootstrap 5.


## Development

```bash
# Install backend dependencies
cd backend && npm install
# Start the server
npm start
```
Environment variables can be customized in `backend/.env`.

Open `http://localhost:3001` in a browser to use the GUI.

### Available API routes

- `GET /api/listings`, `POST /api/listings`
- `GET /api/deals`, `POST /api/deals`
- `GET /api/storage/pins`, `POST /api/storage/pin`



