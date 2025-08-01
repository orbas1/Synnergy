const dotenv = require('dotenv');
const path = require('path');

dotenv.config();

module.exports = {
  port: process.env.PORT || 3001,
  nodeEnv: process.env.NODE_ENV || 'development',
  dataPath: process.env.DATA_PATH || path.join(__dirname, '../services/data.json')
};
