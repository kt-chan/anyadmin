const appConfig = require('../config/app.config');

const logger = {
  info: (message, meta = {}) => {
    if (appConfig.env !== 'test') {
      console.log(`[INFO] ${new Date().toISOString()}: ${message}`, meta);
    }
  },
  error: (message, error) => {
    console.error(`[ERROR] ${new Date().toISOString()}: ${message}`, error);
  },
  warn: (message, meta = {}) => {
    console.warn(`[WARN] ${new Date().toISOString()}: ${message}`, meta);
  },
  debug: (message, meta = {}) => {
    if (appConfig.env === 'development') {
      console.debug(`[DEBUG] ${new Date().toISOString()}: ${message}`, meta);
    }
  }
};

module.exports = logger;
