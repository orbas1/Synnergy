/**
 * Simple request logger middleware.
 * Logs the HTTP method and URL with an ISO timestamp.
 */
function logger(req, res, next) {
  console.log(`[${new Date().toISOString()}] ${req.method} ${req.url}`);
  next();
}

module.exports = logger;
