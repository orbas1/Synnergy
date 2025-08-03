const dotenv = require("dotenv");
const path = require("path");

// Load environment variables from the local .env file regardless of where the
// server is executed from. Using an absolute path ensures consistent behavior
// in production environments where the working directory may differ.
dotenv.config({ path: path.resolve(__dirname, "../.env") });

module.exports = {
  // Normalize the port value to an integer to avoid type issues when the value
  // is supplied via the environment.
  port: parseInt(process.env.PORT, 10) || 3001,
  nodeEnv: process.env.NODE_ENV || "development",
  // Resolve DATA_PATH to an absolute path so that relative values defined in
  // the environment work regardless of the process's working directory.
  dataPath: process.env.DATA_PATH
    ? path.resolve(process.env.DATA_PATH)
    : path.join(__dirname, "../services/data.json"),
};
