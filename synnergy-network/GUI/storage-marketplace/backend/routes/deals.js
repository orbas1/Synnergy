const router = require('express').Router();
const controller = require('../controllers/dealsController');

router.get('/', controller.getDeals);
router.post('/', controller.createDeal);

module.exports = router;
