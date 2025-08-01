const express = require('express');
const router = express.Router();
const ctrl = require('../controllers/contractController');

router.get('/', ctrl.list);
router.post('/', ctrl.deploy);
router.get('/:id', ctrl.get);

module.exports = router;
