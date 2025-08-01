const router = require('express').Router();
const controller = require('../controllers/dealsController');

router.get('/', controller.getDeals);
router.post('/', controller.createDeal);
router.get('/:id', controller.getDeal);

module.exports = router;
