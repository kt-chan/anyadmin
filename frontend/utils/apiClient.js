const axios = require('axios');
const appConfig = require('../config/app.config');

const apiClient = axios.create({
  baseURL: appConfig.backendApiUrl,
  timeout: 5000,
  headers: {
    'Content-Type': 'application/json',
  },
  proxy: false
});

module.exports = apiClient;
