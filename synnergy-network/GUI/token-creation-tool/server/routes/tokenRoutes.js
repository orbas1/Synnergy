import { Router } from 'express';
import { create, index, deploy } from '../controllers/tokenController.js';

const router = Router();
router.get('/', index);
router.post('/', create);
router.post('/deploy', deploy);
export default router;
