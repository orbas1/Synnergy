const express = require("express");
const ctrl = require("../controllers/proposalController");

const router = express.Router();

router.get("/", ctrl.list);
router.post("/", ctrl.submit);
router.get("/:id", ctrl.get);
router.post("/:id/vote", ctrl.vote);
router.post("/:id/execute", ctrl.execute);
router.get("/:id/balance/:asset", ctrl.balance);

module.exports = router;
