const router = require('express').Router();
const controller = require('../controllers/storageController');

router.post('/pin', controller.pin);
router.get('/pins', controller.listPins);
router.get('/retrieve/:cid', controller.retrieve);
router.get('/exists/:cid', controller.exists);
router.post('/create', controller.createStorage);
router.get('/storages', controller.listStorages);

module.exports = router;
