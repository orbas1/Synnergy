const express = require('express');
const morgan = require('morgan');
const contractsRouter = require('./routes/contracts');
const logger = require('./middleware/logger');
require('dotenv').config();

const app = express();
app.use(express.json());
app.use(morgan('dev'));
app.use(logger);

app.use('/api/contracts', contractsRouter);

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Marketplace server listening on port ${PORT}`);
});
