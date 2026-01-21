const apiClient = require('../utils/apiClient');
const logger = require('../utils/logger');

const authService = {
  authenticate: async (username, password) => {
    logger.debug(`Attempting login for user: ${username}`);
    try {
      const response = await apiClient.post('/auth/login', { username, password });
      const user = response.data;
      if (user) {
        logger.info(`User logged in successfully: ${username}`);
        return user;
      }
    } catch (error) {
       logger.warn(`Login failed for user: ${username} - ${error.response?.status} ${error.response?.statusText}`);
    }
    return null;
  }
};

module.exports = authService;
