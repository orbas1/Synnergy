const express = require('express');
const morgan = require('morgan');
const cors = require('cors');
const path = require('path');
const { port } = require('./config');
const listingsRoutes = require('./routes/listings');
const dealsRoutes = require('./routes/deals');
const storageRoutes = require('./routes/storage');

const { errorHandler } = require('./middleware/errorHandler');

const app = express();

app.use(cors());
app.use(morgan('dev'));
app.use(express.json());
app.use(express.static(path.join(__dirname, '..')));

app.use('/api/listings', listingsRoutes);
app.use('/api/deals', dealsRoutes);
app.use('/api/storage', storageRoutes);


app.use(errorHandler);

app.listen(port, () => {
  console.log(`Storage Marketplace API running on port ${port}`);
});
