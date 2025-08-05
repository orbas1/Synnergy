const { env } = require("../config/config");

function errorHandler(err, req, res, next) {
  if (env !== "production") {
    console.error(err.stack);
  }
  res
    .status(err.status || 500)
    .json({ message: err.message || "Internal Server Error" });
}

module.exports = errorHandler;
