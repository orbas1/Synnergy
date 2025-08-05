const express = require("express");
const controller = require("../controllers/nftController");
const router = express.Router();

// Routes for NFT marketplace listing operations
router.get("/listings", controller.all);
router.post("/listings", controller.create);
router.post("/listings/:id/buy", controller.buy);

module.exports = router;
