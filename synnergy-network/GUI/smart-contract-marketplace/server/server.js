const express = require("express");
const morgan = require("morgan");
const contractsRouter = require("./routes/contracts");
const logger = require("./middleware/logger");
const path = require("path");
require("dotenv").config();

const app = express();
app.use(express.json());
app.use(morgan("dev"));
app.use(logger);
app.use(express.static(path.join(__dirname, "..")));

app.use("/api/contracts", contractsRouter);

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Marketplace server listening on port ${PORT}`);
});
