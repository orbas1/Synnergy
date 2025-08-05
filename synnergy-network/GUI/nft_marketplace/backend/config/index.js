const path = require("path");
// Load environment variables from the local .env file
require("dotenv").config({ path: path.join(__dirname, "..", ".env") });

module.exports = {
  // Port the HTTP server listens on
  port: process.env.PORT || 4000,
};
