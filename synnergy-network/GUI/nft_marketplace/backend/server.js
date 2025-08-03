const express = require("express");
const { port } = require("./config");
const logger = require("./middleware/logger");
const nftRoutes = require("./routes/nftRoutes");

const app = express();
app.use(express.json());
app.use(logger);
app.use("/api", nftRoutes);

// Basic error handler
app.use((err, req, res, next) => {
  console.error(err);
  res.status(500).json({ error: "Internal Server Error" });
});

app.listen(port, () => {
  console.log(`NFT Marketplace backend running on port ${port}`);
});
