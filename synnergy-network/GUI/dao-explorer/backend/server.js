const express = require('express');
const bodyParser = require('body-parser');
const cors = require('cors');
const config = require('./config/config');
const proposalRoutes = require('./routes/proposalRoutes');
const errorHandler = require('./middleware/errorHandler');

const app = express();
app.use(cors());
app.use(bodyParser.json());

app.use('/backend/api/proposals', proposalRoutes);
app.use(errorHandler);

app.listen(config.port, () => {
  console.log(`DAO Explorer API running on port ${config.port}`);
});
