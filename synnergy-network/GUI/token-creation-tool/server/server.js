import express from 'express';
import config from 'config';
import dotenv from 'dotenv';
import logger from './middleware/logger.js';
import tokenRoutes from './routes/tokenRoutes.js';

dotenv.config();
const app = express();
app.use(express.json());
app.use(logger);
app.use('/api/tokens', tokenRoutes);

const port = process.env.PORT || config.get('port');
app.listen(port, () => console.log(`Server running on port ${port}`));
