import { Router } from 'express';
import { create } from '../controllers/tokenController.js';

const router = Router();
router.post('/', create);
export default router;
