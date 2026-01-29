const fs = require('fs');
const path = require('path');
const appConfig = require('../config/app.config');

const logDir = path.join(__dirname, '..', '..', 'logs');
const logFile = path.join(logDir, 'frontend.log');

// Ensure log directory exists
if (!fs.existsSync(logDir)) {
  fs.mkdirSync(logDir, { recursive: true });
}

function writeToFile(level, message, meta) {
  let metaStr = '';
  if (meta) {
    if (meta instanceof Error) {
      metaStr = ` ${meta.stack}`;
    } else if (typeof meta === 'object' && Object.keys(meta).length > 0) {
      try {
        metaStr = ` ${JSON.stringify(meta)}`;
      } catch (e) {
        metaStr = ` [Complex Object]`;
      }
    } else if (typeof meta !== 'object') {
      metaStr = ` ${meta}`;
    }
  }
  const logEntry = `[${level}] ${new Date().toISOString()}: ${message}${metaStr}\n`;
  try {
    fs.appendFileSync(logFile, logEntry);
  } catch (err) {
    process.stderr.write(`Failed to write to frontend log file: ${err.message}\n`);
  }
}

const logger = {
  info: (message, meta = {}) => {
    writeToFile('INFO', message, meta);
    if (appConfig.env !== 'test') {
      console.log(`[INFO] ${new Date().toISOString()}: ${message}`, meta);
    }
  },
  error: (message, error) => {
    writeToFile('ERROR', message, error);
    console.error(`[ERROR] ${new Date().toISOString()}: ${message}`, error);
  },
  warn: (message, meta = {}) => {
    writeToFile('WARN', message, meta);
    console.warn(`[WARN] ${new Date().toISOString()}: ${message}`, meta);
  },
  debug: (message, meta = {}) => {
    writeToFile('DEBUG', message, meta);
    if (appConfig.env === 'development') {
      console.debug(`[DEBUG] ${new Date().toISOString()}: ${message}`, meta);
    }
  }
};

module.exports = logger;
