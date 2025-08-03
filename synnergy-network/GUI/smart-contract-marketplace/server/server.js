import express from "express";
import morgan from "morgan";
import contractsRouter from "./routes/contracts.js";
import logger from "./middleware/logger.js";
import path from "path";
import dotenv from "dotenv";
import { fileURLToPath } from "url";

dotenv.config();

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

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
