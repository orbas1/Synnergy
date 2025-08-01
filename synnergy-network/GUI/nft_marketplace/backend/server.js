const express = require('express');
const bodyParser = require('body-parser');
const { port } = require('./config');
const logger = require('./middleware/logger');
const nftRoutes = require('./routes/nftRoutes');

const app = express();
app.use(bodyParser.json());
app.use(logger);
app.use('/api', nftRoutes);

app.listen(port, () => {
  console.log(`NFT Marketplace backend running on port ${port}`);
});
