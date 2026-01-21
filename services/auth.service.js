const { users } = require('../data/mockData');
const logger = require('../utils/logger');

const authService = {
  authenticate: async (username, password) => {
    logger.debug(`Attempting login for user: ${username}`);
    const user = users.find(u => u.username === username && u.password === password);
    
    if (user) {
      logger.info(`User logged in successfully: ${username}`);
      return {
        username: user.username,
        role: user.role
      };
    }
    
    logger.warn(`Login failed for user: ${username}`);
    return null;
  }
};

module.exports = authService;
