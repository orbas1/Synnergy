const router = require('express').Router();
const controller = require('../controllers/listingsController');

router.get('/', controller.getListings);
router.post('/', controller.createListing);

module.exports = router;
