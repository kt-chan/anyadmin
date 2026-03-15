const axios = require('axios');
const appConfig = require('../config/app.config');
const logger = require('./logger');

const apiClient = axios.create({
  baseURL: appConfig.backendApiUrl,
  timeout: 300000,
  headers: {
    'Content-Type': 'application/json',
  },
  proxy: false
});

// Add a response interceptor
apiClient.interceptors.response.use(function (response) {
  return response;
}, function (error) {
  if (error.response) {
    // The request was made and the server responded with a status code
    // that falls out of the range of 2xx
    logger.error('Backend API Error:', {
      status: error.response.status,
      data: error.response.data,
      url: error.config.url,
      method: error.config.method
    });
  } else if (error.request) {
    // The request was made but no response was received
    logger.error('Backend API No Response:', {
      url: error.config.url,
      method: error.config.method
    });
  } else {
    // Something happened in setting up the request that triggered an Error
    logger.error('Backend API Request Setup Error:', error.message);
  }
  return Promise.reject(error);
});

module.exports = apiClient;
